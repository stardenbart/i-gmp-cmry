package masterusecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/masterimportxlsx"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const importTokenPrefix = "master-import:validation:"

var errImportTokenInvalid = errors.New("validation token expired or used")

type MasterImportError struct {
	Status  int
	Message string
	Data    any
}

func (e *MasterImportError) Error() string { return e.Message }

type MasterImportUseCase struct {
	db      *gorm.DB
	redis   *redis.Client
	heiRepo master.HEIRepository
}

func NewMasterImportUseCase(db *gorm.DB, rdb *redis.Client, heiRepo master.HEIRepository) *MasterImportUseCase {
	return &MasterImportUseCase{db: db, redis: rdb, heiRepo: heiRepo}
}

func (uc *MasterImportUseCase) CheckRateLimit(ctx context.Context, userID string) error {
	if uc.redis == nil {
		return importErr(http.StatusServiceUnavailable, "Redis tidak tersedia; import dinonaktifkan sementara", nil)
	}
	key := "master-import:rate:" + userID
	count, err := uc.redis.Eval(ctx, `
		local count = redis.call('INCR', KEYS[1])
		if count == 1 then redis.call('EXPIRE', KEYS[1], ARGV[1]) end
		return count`, []string{key}, int((10*time.Minute)/time.Second)).Int64()
	if err != nil {
		return importErr(http.StatusServiceUnavailable, "Redis tidak tersedia; import dinonaktifkan sementara", nil)
	}
	if count > 5 {
		return importErr(http.StatusTooManyRequests, "batas lima validasi import per sepuluh menit telah tercapai", nil)
	}
	return nil
}

func IsMasterImportType(value string) bool {
	switch value {
	case master.ImportTypeAspek, master.ImportTypeDetail, master.ImportTypeUraian, master.ImportTypeHEI:
		return true
	default:
		return false
	}
}

func (uc *MasterImportUseCase) GenerateTemplate(ctx context.Context, importType string, actor master.ImportActor, requestedPlant string) ([]byte, string, error) {
	if !IsMasterImportType(importType) {
		return nil, "", importErr(http.StatusBadRequest, "jenis import tidak didukung", nil)
	}
	plantID, err := resolveImportPlant(importType, actor, requestedPlant)
	if err != nil {
		return nil, "", err
	}

	var headers []string
	var references []masterimportxlsx.ReferenceRow
	switch importType {
	case master.ImportTypeAspek:
		headers = []string{"area_id", "area_name", "plant_id"}
		var rows []struct{ AreaID, AreaName, PlantID string }
		if err := uc.db.WithContext(ctx).Raw(`
			SELECT "AreaID" AS area_id, "AreaName" AS area_name, coalesce("PlantID", '') AS plant_id
			FROM "Area_Master" WHERE "PlantID" = ? ORDER BY "AreaName"`, plantID).Scan(&rows).Error; err != nil {
			return nil, "", err
		}
		for _, row := range rows {
			references = append(references, masterimportxlsx.ReferenceRow{row.AreaID, row.AreaName, row.PlantID})
		}
	case master.ImportTypeDetail:
		headers = []string{"aspek_id", "aspek_name", "area_name", "plant_id"}
		var rows []struct{ AspekID, AspekName, AreaName, PlantID string }
		if err := uc.db.WithContext(ctx).Raw(`
			SELECT asp."AspekID" AS aspek_id, asp."AspekName" AS aspek_name,
			       area."AreaName" AS area_name, coalesce(area."PlantID", '') AS plant_id
			FROM "Aspek_Master" asp
			JOIN "Area_Master" area ON area."AreaID" = asp."AreaID"
			WHERE area."PlantID" = ? ORDER BY area."AreaName", asp."AspekName"`, plantID).Scan(&rows).Error; err != nil {
			return nil, "", err
		}
		for _, row := range rows {
			references = append(references, masterimportxlsx.ReferenceRow{row.AspekID, row.AspekName, row.AreaName, row.PlantID})
		}
	case master.ImportTypeUraian:
		headers = []string{"detail_id", "detail_name", "aspek_name", "area_name", "plant_id"}
		var rows []struct{ DetailID, DetailName, AspekName, AreaName, PlantID string }
		if err := uc.db.WithContext(ctx).Raw(`
			SELECT det."DetailID" AS detail_id, det."DetailName" AS detail_name,
			       asp."AspekName" AS aspek_name, area."AreaName" AS area_name,
			       coalesce(area."PlantID", '') AS plant_id
			FROM "Detail_Master" det
			JOIN "Aspek_Master" asp ON asp."AspekID" = det."AspekID"
			JOIN "Area_Master" area ON area."AreaID" = asp."AreaID"
			WHERE area."PlantID" = ? ORDER BY area."AreaName", asp."AspekName", det."DetailName"`, plantID).Scan(&rows).Error; err != nil {
			return nil, "", err
		}
		for _, row := range rows {
			references = append(references, masterimportxlsx.ReferenceRow{row.DetailID, row.DetailName, row.AspekName, row.AreaName, row.PlantID})
		}
	case master.ImportTypeHEI:
		headers = []string{"existing_category_name_optional"}
		categories, err := uc.heiRepo.FindCategories()
		if err != nil {
			return nil, "", err
		}
		for _, category := range categories {
			references = append(references, masterimportxlsx.ReferenceRow{category})
		}
	}

	content, err := masterimportxlsx.Generate(importType, plantID, headers, references)
	if err != nil {
		return nil, "", err
	}
	return content, fmt.Sprintf("template-import-%s-%s.xlsx", importType, time.Now().Format("20060102-150405")), nil
}

func (uc *MasterImportUseCase) Validate(ctx context.Context, importType string, actor master.ImportActor, requestedPlant, fileName string, content []byte) (*master.ImportPreview, error) {
	if !IsMasterImportType(importType) {
		return nil, importErr(http.StatusBadRequest, "jenis import tidak didukung", nil)
	}
	plantID, err := resolveImportPlant(importType, actor, requestedPlant)
	if err != nil {
		return nil, err
	}
	uc.expireStaleBatches(ctx)
	fileHash := sha256Hex(content)
	importID := idgen.GenerateRandom("MIMP")
	batch := &master.MasterImportBatch{
		ImportID: importID, UserID: actor.UserID, PlantID: nullableString(plantID),
		ImportType: importType, OriginalFileName: sanitizeFileName(fileName), FileHash: fileHash,
		TemplateVersion: master.ImportTemplateV1, Status: "Uploaded",
	}
	if err := uc.db.WithContext(ctx).Create(batch).Error; err != nil {
		return nil, importErr(http.StatusInternalServerError, "gagal membuat catatan audit import", nil)
	}
	var previousCommits int64
	duplicateQuery := uc.db.WithContext(ctx).Model(&master.MasterImportBatch{}).
		Where(`"ImportID" <> ? AND "FileHash" = ? AND "ImportType" = ? AND "Status" = 'Committed'`, importID, fileHash, importType)
	if plantID == "" {
		duplicateQuery = duplicateQuery.Where(`"PlantID" IS NULL`)
	} else {
		duplicateQuery = duplicateQuery.Where(`"PlantID" = ?`, plantID)
	}
	if err := duplicateQuery.Count(&previousCommits).Error; err != nil {
		uc.markBatchFailed(ctx, importID, "Failed", "duplicate import audit check failed")
		return nil, importErr(http.StatusInternalServerError, "gagal memeriksa riwayat import", nil)
	}
	if previousCommits > 0 {
		message := "file yang sama sudah pernah berhasil diimport untuk jenis dan plant ini"
		uc.markBatchFailed(ctx, importID, "ValidationRejected", message)
		return nil, importErr(http.StatusConflict, message, nil)
	}

	parsed, err := masterimportxlsx.Parse(fileName, content, importType)
	if err != nil {
		uc.markBatchFailed(ctx, importID, "ValidationRejected", err.Error())
		return nil, importErr(http.StatusUnprocessableEntity, err.Error(), nil)
	}
	if importType != master.ImportTypeHEI && parsed.PlantID != plantID {
		message := "plant template tidak sesuai; unduh template terbaru untuk plant yang dipilih"
		uc.markBatchFailed(ctx, importID, "ValidationRejected", message)
		return nil, importErr(http.StatusForbidden, message, nil)
	}

	rows, summary, normalizedHash, err := uc.validateParsedRows(ctx, uc.db, importType, plantID, parsed)
	if err != nil {
		uc.markBatchFailed(ctx, importID, "Failed", "internal validation error")
		return nil, err
	}
	now := time.Now()
	status := "Validated"
	if summary.Invalid > 0 {
		status = "ValidationRejected"
	}
	_ = uc.db.WithContext(ctx).Model(&master.MasterImportBatch{}).Where(`"ImportID" = ?`, importID).Updates(map[string]any{
		"TotalRows": len(rows), "ValidRows": summary.Valid, "InvalidRows": summary.Invalid,
		"Status": status, "ValidatedAt": now, "UpdatedAt": now,
	}).Error

	preview := &master.ImportPreview{
		ImportID: importID, Type: importType, TemplateVersion: parsed.TemplateVersion,
		FileHash: fileHash, ReferenceAt: parsed.ReferenceAt, Summary: summary, Rows: rows,
		CanCommit: summary.Invalid == 0,
	}
	if !preview.CanCommit {
		return preview, importErr(http.StatusUnprocessableEntity, "file mengandung data tidak valid atau duplikat", preview)
	}

	token, err := uc.storeValidationToken(ctx, master.ImportTokenState{
		ImportID: importID, UserID: actor.UserID, PlantID: plantID, ImportType: importType,
		FileHash: fileHash, TemplateVersion: parsed.TemplateVersion, RowCount: len(rows),
		NormalizedRowsHash: normalizedHash, Summary: summary, Status: "validated",
		ExpiresAt: time.Now().Add(master.ImportTokenTTL),
	})
	if err != nil {
		uc.markBatchFailed(ctx, importID, "Failed", "validation token store unavailable")
		return nil, importErr(http.StatusServiceUnavailable, "Redis tidak tersedia; validasi import tidak dapat disimpan", nil)
	}
	preview.ValidationToken = token
	return preview, nil
}

func (uc *MasterImportUseCase) Commit(ctx context.Context, importType string, actor master.ImportActor, requestedPlant, token, fileName string, content []byte) (*master.ImportCommitResult, error) {
	if !IsMasterImportType(importType) {
		return nil, importErr(http.StatusBadRequest, "jenis import tidak didukung", nil)
	}
	plantID, err := resolveImportPlant(importType, actor, requestedPlant)
	if err != nil {
		return nil, err
	}
	state, tokenKey, err := uc.acquireValidationToken(ctx, token)
	if err != nil {
		return nil, err
	}
	defer uc.redis.Del(context.Background(), tokenKey)

	if state.UserID != actor.UserID || state.PlantID != plantID || state.ImportType != importType {
		uc.markBatchFailed(ctx, state.ImportID, "Failed", "validation token scope mismatch")
		return nil, importErr(http.StatusForbidden, "token validasi tidak sesuai dengan pengguna, plant, atau jenis import", nil)
	}
	if state.FileHash != sha256Hex(content) {
		uc.markBatchFailed(ctx, state.ImportID, "Failed", "file hash mismatch")
		return nil, importErr(http.StatusConflict, "file berubah setelah validasi; lakukan validasi ulang", nil)
	}

	parsed, err := masterimportxlsx.Parse(fileName, content, importType)
	if err != nil {
		uc.markBatchFailed(ctx, state.ImportID, "Failed", err.Error())
		return nil, importErr(http.StatusUnprocessableEntity, err.Error(), nil)
	}
	if parsed.TemplateVersion != state.TemplateVersion || (importType != master.ImportTypeHEI && parsed.PlantID != plantID) {
		uc.markBatchFailed(ctx, state.ImportID, "Failed", "template metadata mismatch")
		return nil, importErr(http.StatusConflict, "metadata template berubah setelah validasi", nil)
	}

	now := time.Now()
	if err := uc.db.WithContext(ctx).Model(&master.MasterImportBatch{}).Where(`"ImportID" = ?`, state.ImportID).
		Updates(map[string]any{"Status": "Committing", "UpdatedAt": now}).Error; err != nil {
		return nil, importErr(http.StatusInternalServerError, "gagal memperbarui audit import", nil)
	}

	result := &master.ImportCommitResult{ImportID: state.ImportID, Type: importType}
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockScope := "master-hei:global"
		if importType != master.ImportTypeHEI {
			lockScope = "master-audit-tree:" + plantID
		}
		if lockErr := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?, 0))`, lockScope).Error; lockErr != nil {
			return lockErr
		}

		validatedRows, summary, normalizedHash, validateErr := uc.validateParsedRows(ctx, tx, importType, plantID, parsed)
		if validateErr != nil {
			return validateErr
		}
		if summary.Invalid > 0 || normalizedHash != state.NormalizedRowsHash || len(validatedRows) != state.RowCount {
			return importErr(http.StatusConflict, "data atau referensi berubah setelah preview; lakukan validasi ulang", &master.ImportPreview{
				ImportID: state.ImportID, Type: importType, TemplateVersion: parsed.TemplateVersion,
				FileHash: state.FileHash, Summary: summary, Rows: validatedRows, CanCommit: false,
			})
		}

		createdRows, insertErr := insertMasterRows(tx, state.ImportID, importType, validatedRows)
		if insertErr != nil {
			if isUniqueViolation(insertErr) {
				return importErr(http.StatusConflict, "data duplikat terdeteksi saat commit", nil)
			}
			return insertErr
		}
		result.Inserted = len(createdRows)
		result.CreatedRows = createdRows
		result.NextImportType = nextImportType(importType)
		committedAt := time.Now()
		return tx.Model(&master.MasterImportBatch{}).Where(`"ImportID" = ?`, state.ImportID).Updates(map[string]any{
			"Status": "Committed", "InsertedRows": len(createdRows), "CommittedAt": committedAt, "UpdatedAt": committedAt,
		}).Error
	})
	if err != nil {
		uc.markBatchFailed(context.Background(), state.ImportID, "Failed", err.Error())
		var importError *MasterImportError
		if errors.As(err, &importError) {
			return nil, importError
		}
		return nil, importErr(http.StatusInternalServerError, "commit import gagal dan seluruh data dibatalkan", nil)
	}
	if importType == master.ImportTypeHEI {
		uc.heiRepo.InvalidateCache()
	}
	return result, nil
}

func (uc *MasterImportUseCase) validateParsedRows(ctx context.Context, db *gorm.DB, importType, plantID string, parsed *masterimportxlsx.ParsedWorkbook) ([]master.ImportPreviewRow, master.ImportSummary, string, error) {
	normalized, err := normalizeValues(ctx, db, parsed.Rows)
	if err != nil {
		return nil, master.ImportSummary{}, "", importErr(http.StatusInternalServerError, "fungsi normalisasi database belum tersedia; jalankan migration", nil)
	}

	rows := make([]master.ImportPreviewRow, 0, len(parsed.Rows))
	keys := make([][]string, 0, len(parsed.Rows))
	for _, source := range parsed.Rows {
		preview := master.ImportPreviewRow{Row: source.SourceRow, Status: "valid", NormalizedData: map[string]any{}, Errors: []master.ImportFieldError{}}
		for field, raw := range source.Values {
			value := normalized[raw]
			cleaned := value.Cleaned
			if field == "hei_code" {
				cleaned = value.Code
			}
			preview.NormalizedData[field] = cleaned
		}
		validateFieldValues(importType, &preview)
		rows = append(rows, preview)
		keys = append(keys, duplicateKeys(importType, preview.NormalizedData, normalized, source.Values))
	}

	if err := uc.validateParents(ctx, db, importType, plantID, rows, normalized, parsed.Rows); err != nil {
		return nil, master.ImportSummary{}, "", err
	}

	groups := make(map[string][]int)
	for i, rowKeys := range keys {
		if len(rows[i].Errors) != 0 {
			continue
		}
		for _, key := range rowKeys {
			if key != "" {
				groups[key] = append(groups[key], i)
			}
		}
	}
	for _, indexes := range groups {
		if len(indexes) > 1 {
			for _, index := range indexes {
				addRowError(&rows[index], "", "DUPLICATE_IN_FILE", "data yang sama muncul lebih dari sekali dalam file")
			}
		}
	}

	existing, err := existingDuplicateKeys(ctx, db, importType, plantID)
	if err != nil {
		return nil, master.ImportSummary{}, "", err
	}
	for i, rowKeys := range keys {
		for _, key := range rowKeys {
			if key != "" && existing[key] {
				addRowError(&rows[i], "", "DUPLICATE_IN_DATABASE", "data yang sama sudah tersedia di database")
				break
			}
		}
	}

	summary := master.ImportSummary{Total: len(rows)}
	for i := range rows {
		if len(rows[i].Errors) > 0 {
			rows[i].Status = "invalid"
			summary.Invalid++
		} else {
			summary.Valid++
		}
		if hasErrorCode(rows[i], "DUPLICATE_IN_FILE") {
			summary.DuplicateInFile++
		}
		if hasErrorCode(rows[i], "DUPLICATE_IN_DATABASE") {
			summary.DuplicateInDatabase++
		}
	}
	hashPayload, _ := json.Marshal(rows)
	return rows, summary, sha256Hex(hashPayload), nil
}

type normalizedValue struct {
	Cleaned    string
	Normalized string
	Code       string
}

func normalizeValues(ctx context.Context, db *gorm.DB, rows []masterimportxlsx.ParsedRow) (map[string]normalizedValue, error) {
	unique := map[string]bool{}
	for _, row := range rows {
		for _, raw := range row.Values {
			unique[raw] = true
		}
	}
	values := make([]string, 0, len(unique))
	for value := range unique {
		values = append(values, value)
	}
	sort.Strings(values)
	payload, _ := json.Marshal(values)
	var results []struct {
		Raw, Cleaned, Normalized, Code string
	}
	err := db.WithContext(ctx).Raw(`
		SELECT item.value AS raw,
		       coalesce(master_clean_text(item.value), '') AS cleaned,
		       coalesce(master_normalize_text(item.value), '') AS normalized,
		       coalesce(master_normalize_code(item.value), '') AS code
		FROM jsonb_array_elements_text(?::jsonb) AS item(value)`, string(payload)).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	output := make(map[string]normalizedValue, len(results))
	for _, result := range results {
		output[result.Raw] = normalizedValue{Cleaned: result.Cleaned, Normalized: result.Normalized, Code: result.Code}
	}
	return output, nil
}

func validateFieldValues(importType string, row *master.ImportPreviewRow) {
	required := func(field string, max int) string {
		value, _ := row.NormalizedData[field].(string)
		if value == "" {
			addRowError(row, field, "REQUIRED", "kolom wajib diisi")
		} else if len([]rune(value)) > max {
			addRowError(row, field, "TOO_LONG", fmt.Sprintf("maksimal %d karakter", max))
		}
		return value
	}
	optional := func(field string, max int) string {
		value, _ := row.NormalizedData[field].(string)
		if len([]rune(value)) > max {
			addRowError(row, field, "TOO_LONG", fmt.Sprintf("maksimal %d karakter", max))
		}
		return value
	}

	switch importType {
	case master.ImportTypeAspek:
		required("area_id", 50)
		required("aspek_name", 100)
	case master.ImportTypeDetail:
		required("aspek_id", 50)
		required("detail_name", 150)
	case master.ImportTypeUraian:
		required("detail_id", 50)
		required("uraian_text", 500)
		scoreRaw := required("standard_score", 3)
		score, err := strconv.Atoi(scoreRaw)
		if err != nil || score < 1 || score > 100 {
			addRowError(row, "standard_score", "INVALID_VALUE", "skor standar harus bilangan 1 sampai 100")
		} else {
			row.NormalizedData["standard_score"] = score
		}
	case master.ImportTypeHEI:
		required("category_name", 50)
		optional("hei_code", 50)
		required("hei_name", 150)
		optional("description", 255)
		status := optional("status", 20)
		if status == "" {
			status = "Active"
		}
		if !strings.EqualFold(status, "Active") && !strings.EqualFold(status, "Inactive") {
			addRowError(row, "status", "INVALID_VALUE", "status hanya boleh Active atau Inactive")
		} else if strings.EqualFold(status, "Active") {
			row.NormalizedData["status"] = "Active"
		} else {
			row.NormalizedData["status"] = "Inactive"
		}
	}
}

func duplicateKeys(importType string, data map[string]any, normalized map[string]normalizedValue, raw map[string]string) []string {
	value := func(field string) string { return normalized[raw[field]].Normalized }
	code := func(field string) string { return normalized[raw[field]].Code }
	switch importType {
	case master.ImportTypeAspek:
		return []string{fmt.Sprintf("%s\x00%s", data["area_id"], value("aspek_name"))}
	case master.ImportTypeDetail:
		return []string{fmt.Sprintf("%s\x00%s", data["aspek_id"], value("detail_name"))}
	case master.ImportTypeUraian:
		return []string{fmt.Sprintf("%s\x00%s", data["detail_id"], value("uraian_text"))}
	case master.ImportTypeHEI:
		nameKey := fmt.Sprintf("name:%s\x00%s", value("category_name"), value("hei_name"))
		if code("hei_code") != "" {
			return []string{nameKey, "code:" + value("category_name") + "\x00" + code("hei_code")}
		}
		return []string{nameKey}
	default:
		return nil
	}
}

func (uc *MasterImportUseCase) validateParents(ctx context.Context, db *gorm.DB, importType, plantID string, rows []master.ImportPreviewRow, normalized map[string]normalizedValue, sourceRows []masterimportxlsx.ParsedRow) error {
	if importType == master.ImportTypeHEI {
		// CategoryName is a value on HEI_Master, not a foreign key. Existing
		// categories in the template are suggestions only; importing a new
		// category is valid and the first HEI row creates it implicitly.
		return nil
	}

	field := map[string]string{master.ImportTypeAspek: "area_id", master.ImportTypeDetail: "aspek_id", master.ImportTypeUraian: "detail_id"}[importType]
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if id, _ := row.NormalizedData[field].(string); id != "" {
			ids = append(ids, id)
		}
	}
	valid := map[string]bool{}
	var found []string
	var query string
	switch importType {
	case master.ImportTypeAspek:
		query = `SELECT area."AreaID" FROM "Area_Master" area WHERE area."PlantID" = ? AND area."AreaID" IN ?`
	case master.ImportTypeDetail:
		query = `SELECT asp."AspekID" FROM "Aspek_Master" asp JOIN "Area_Master" area ON area."AreaID" = asp."AreaID" WHERE area."PlantID" = ? AND asp."AspekID" IN ?`
	case master.ImportTypeUraian:
		query = `SELECT det."DetailID" FROM "Detail_Master" det JOIN "Aspek_Master" asp ON asp."AspekID" = det."AspekID" JOIN "Area_Master" area ON area."AreaID" = asp."AreaID" WHERE area."PlantID" = ? AND det."DetailID" IN ?`
	}
	if len(ids) > 0 {
		if err := db.WithContext(ctx).Raw(query, plantID, ids).Scan(&found).Error; err != nil {
			return err
		}
	}
	for _, id := range found {
		valid[id] = true
	}
	for i := range rows {
		id, _ := rows[i].NormalizedData[field].(string)
		if id != "" && !valid[id] {
			addRowError(&rows[i], field, "INVALID_REFERENCE", "parent tidak ditemukan atau berada di plant lain")
		}
	}
	return nil
}

func existingDuplicateKeys(ctx context.Context, db *gorm.DB, importType, plantID string) (map[string]bool, error) {
	keys := map[string]bool{}
	switch importType {
	case master.ImportTypeAspek:
		var rows []struct{ Parent, Name string }
		err := db.WithContext(ctx).Raw(`SELECT asp."AreaID" AS parent, master_normalize_text(asp."AspekName") AS name FROM "Aspek_Master" asp JOIN "Area_Master" area ON area."AreaID" = asp."AreaID" WHERE area."PlantID" = ?`, plantID).Scan(&rows).Error
		for _, row := range rows {
			keys[row.Parent+"\x00"+row.Name] = true
		}
		return keys, err
	case master.ImportTypeDetail:
		var rows []struct{ Parent, Name string }
		err := db.WithContext(ctx).Raw(`SELECT det."AspekID" AS parent, master_normalize_text(det."DetailName") AS name FROM "Detail_Master" det JOIN "Aspek_Master" asp ON asp."AspekID" = det."AspekID" JOIN "Area_Master" area ON area."AreaID" = asp."AreaID" WHERE area."PlantID" = ?`, plantID).Scan(&rows).Error
		for _, row := range rows {
			keys[row.Parent+"\x00"+row.Name] = true
		}
		return keys, err
	case master.ImportTypeUraian:
		var rows []struct{ Parent, Name string }
		err := db.WithContext(ctx).Raw(`SELECT ur."DetailID" AS parent, master_normalize_text(ur."UraianText") AS name FROM "Uraian_Master" ur JOIN "Detail_Master" det ON det."DetailID" = ur."DetailID" JOIN "Aspek_Master" asp ON asp."AspekID" = det."AspekID" JOIN "Area_Master" area ON area."AreaID" = asp."AreaID" WHERE area."PlantID" = ?`, plantID).Scan(&rows).Error
		for _, row := range rows {
			keys[row.Parent+"\x00"+row.Name] = true
		}
		return keys, err
	case master.ImportTypeHEI:
		var rows []struct{ Category, Name, Code string }
		err := db.WithContext(ctx).Raw(`SELECT master_normalize_text("CategoryName") AS category, master_normalize_text("HEIName") AS name, coalesce(master_normalize_code("HEICode"), '') AS code FROM "HEI_Master"`).Scan(&rows).Error
		for _, row := range rows {
			keys["name:"+row.Category+"\x00"+row.Name] = true
			if row.Code != "" {
				keys["code:"+row.Category+"\x00"+row.Code] = true
			}
		}
		return keys, err
	}
	return keys, nil
}

func insertMasterRows(tx *gorm.DB, importID, importType string, rows []master.ImportPreviewRow) ([]master.ImportCreatedRow, error) {
	created := make([]master.ImportCreatedRow, 0, len(rows))
	links := make([]master.MasterImportRow, 0, len(rows))
	for _, row := range rows {
		data := row.NormalizedData
		var entityID string
		switch importType {
		case master.ImportTypeAspek:
			entityID = idgen.GenerateRandom("ASP")
			if err := tx.Create(&master.Aspek{AspekID: entityID, AreaID: stringValue(data, "area_id"), AspekName: stringValue(data, "aspek_name")}).Error; err != nil {
				return nil, err
			}
		case master.ImportTypeDetail:
			entityID = idgen.GenerateRandom("DET")
			if err := tx.Create(&master.Detail{DetailID: entityID, AspekID: stringValue(data, "aspek_id"), DetailName: stringValue(data, "detail_name")}).Error; err != nil {
				return nil, err
			}
		case master.ImportTypeUraian:
			entityID = idgen.GenerateRandom("URN")
			if err := tx.Create(&master.Uraian{UraianID: entityID, DetailID: stringValue(data, "detail_id"), UraianText: stringValue(data, "uraian_text"), StandardScore: intValue(data, "standard_score")}).Error; err != nil {
				return nil, err
			}
		case master.ImportTypeHEI:
			entityID = idgen.GenerateRandom("HEI")
			if err := tx.Create(&master.HEIMaster{HEIID: entityID, CategoryName: stringValue(data, "category_name"), HEICode: stringValue(data, "hei_code"), HEIName: stringValue(data, "hei_name"), Description: stringValue(data, "description"), Status: stringValue(data, "status")}).Error; err != nil {
				return nil, err
			}
		}
		links = append(links, master.MasterImportRow{ImportID: importID, EntityType: importType, EntityID: entityID, SourceRow: row.Row})
		createdData := make(map[string]any, len(data)+1)
		for key, value := range data {
			createdData[key] = value
		}
		createdData[entityIDField(importType)] = entityID
		created = append(created, master.ImportCreatedRow{SourceRow: row.Row, EntityID: entityID, Data: createdData})
	}
	if len(links) > 0 {
		if err := tx.CreateInBatches(links, 250).Error; err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (uc *MasterImportUseCase) storeValidationToken(ctx context.Context, state master.ImportTokenState) (string, error) {
	if uc.redis == nil {
		return "", errors.New("redis client unavailable")
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	token := hex.EncodeToString(random)
	payload, _ := json.Marshal(state)
	if err := uc.redis.Set(ctx, importTokenPrefix+sha256Hex([]byte(token)), payload, master.ImportTokenTTL).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (uc *MasterImportUseCase) acquireValidationToken(ctx context.Context, token string) (*master.ImportTokenState, string, error) {
	if uc.redis == nil {
		return nil, "", importErr(http.StatusServiceUnavailable, "Redis tidak tersedia; commit import dinonaktifkan sementara", nil)
	}
	if strings.TrimSpace(token) == "" {
		return nil, "", importErr(http.StatusBadRequest, "token validasi tidak tersedia", nil)
	}
	key := importTokenPrefix + sha256Hex([]byte(token))
	var state master.ImportTokenState
	err := uc.redis.Watch(ctx, func(tx *redis.Tx) error {
		payload, err := tx.Get(ctx, key).Bytes()
		if err != nil {
			return err
		}
		if err := json.Unmarshal(payload, &state); err != nil {
			return err
		}
		if state.Status != "validated" || time.Now().After(state.ExpiresAt) {
			return errImportTokenInvalid
		}
		state.Status = "committing"
		updated, _ := json.Marshal(state)
		ttl, err := tx.TTL(ctx, key).Result()
		if err != nil || ttl <= 0 {
			return errImportTokenInvalid
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error { pipe.Set(ctx, key, updated, ttl); return nil })
		return err
	}, key)
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, errImportTokenInvalid) {
			return nil, key, importErr(http.StatusGone, "token validasi kedaluwarsa atau sudah digunakan", nil)
		}
		if errors.Is(err, redis.TxFailedErr) {
			return nil, key, importErr(http.StatusConflict, "token validasi sedang digunakan oleh request lain", nil)
		}
		return nil, key, importErr(http.StatusServiceUnavailable, "Redis tidak tersedia; status token tidak dapat diverifikasi", nil)
	}
	return &state, key, nil
}

func resolveImportPlant(importType string, actor master.ImportActor, requested string) (string, error) {
	if importType == master.ImportTypeHEI {
		return "", nil
	}
	requested = strings.TrimSpace(requested)
	if actor.PlantID != "" {
		if requested != "" && requested != actor.PlantID {
			return "", importErr(http.StatusForbidden, "plant target tidak sesuai dengan plant pengguna", nil)
		}
		return actor.PlantID, nil
	}
	if actor.RoleID != "ROLE-000" && actor.RoleID != "SUPERADMIN" {
		return "", importErr(http.StatusForbidden, "pengguna tidak memiliki scope plant", nil)
	}
	if requested == "" {
		return "", importErr(http.StatusBadRequest, "Super Admin wajib memilih plant target", nil)
	}
	return requested, nil
}

func (uc *MasterImportUseCase) markBatchFailed(ctx context.Context, importID, status, summary string) {
	if len(summary) > 2000 {
		summary = summary[:2000]
	}
	_ = uc.db.WithContext(ctx).Model(&master.MasterImportBatch{}).Where(`"ImportID" = ?`, importID).Updates(map[string]any{"Status": status, "ErrorSummary": summary, "UpdatedAt": time.Now()}).Error
}

func (uc *MasterImportUseCase) expireStaleBatches(ctx context.Context) {
	cutoff := time.Now().Add(-master.ImportTokenTTL)
	_ = uc.db.WithContext(ctx).Model(&master.MasterImportBatch{}).
		Where(`"Status" IN ('Validated', 'Committing') AND "UpdatedAt" < ?`, cutoff).
		Updates(map[string]any{"Status": "Expired", "ErrorSummary": "validation token expired before commit completed", "UpdatedAt": time.Now()}).Error
}

func importErr(status int, message string, data any) error {
	return &MasterImportError{Status: status, Message: message, Data: data}
}
func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
func sanitizeFileName(name string) string {
	name = filepath.Base(strings.ReplaceAll(strings.TrimSpace(name), "\\", "/"))
	if len(name) > 255 {
		name = name[:255]
	}
	return name
}
func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func addRowError(row *master.ImportPreviewRow, field, code, message string) {
	for _, item := range row.Errors {
		if item.Field == field && item.Code == code {
			return
		}
	}
	row.Errors = append(row.Errors, master.ImportFieldError{Field: field, Code: code, Message: message})
}
func hasErrorCode(row master.ImportPreviewRow, code string) bool {
	for _, item := range row.Errors {
		if item.Code == code {
			return true
		}
	}
	return false
}
func stringValue(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return value
}
func intValue(data map[string]any, key string) int { value, _ := data[key].(int); return value }
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505") || strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
func entityIDField(importType string) string {
	return map[string]string{master.ImportTypeAspek: "aspek_id", master.ImportTypeDetail: "detail_id", master.ImportTypeUraian: "uraian_id", master.ImportTypeHEI: "hei_id"}[importType]
}
func nextImportType(importType string) string {
	return map[string]string{master.ImportTypeAspek: master.ImportTypeDetail, master.ImportTypeDetail: master.ImportTypeUraian}[importType]
}

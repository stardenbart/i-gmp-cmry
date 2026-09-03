package exporter

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	_ "golang.org/x/image/webp"
)

// PlaceholderPayload menyimpan data statis (Header) dan dinamis (Items)
type PlaceholderPayload struct {
	Headers                    map[string]interface{}
	Items                      []map[string]interface{}
	FlattenItemRows            bool
	TrimReservedItemRows       bool
	MergeDuplicateItemFields   []string
	MergeDuplicateWithinFields []string
	FixedCells                 map[string]interface{}
	LogoCell                   string
	LogoPaths                  []string
	RemoveColumns              []string
	TotalRow                   *TableTotalRow
}

// TableTotalRow appends a visible grand-total row immediately after the last
// exported item while preserving any existing template content below it.
type TableTotalRow struct {
	Label            string
	LabelStartColumn string
	LabelEndColumn   string
	ValueColumn      string
	Value            interface{}
}

// GenerateExcelWithPlaceholder adalah engine utama pencarian dan penggantian (search & replace)
func GenerateExcelWithPlaceholder(templatePath, sheetName string, payload *PlaceholderPayload) (*bytes.Buffer, error) {
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka template: %w", err)
	}
	defer f.Close()

	if sheetName == "" {
		sheetName = f.GetSheetName(f.GetActiveSheetIndex())
	}
	for _, column := range payload.RemoveColumns {
		if err := f.RemoveCol(sheetName, column); err != nil {
			return nil, fmt.Errorf("gagal menghapus kolom template %s: %w", column, err)
		}
	}

	// 1. Cari baris mana yang merupakan "Template Row" tabel (mengandung "{item.")
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca baris: %w", err)
	}

	// Build a set of item key placeholders to detect template row
	itemKeySet := make(map[string]bool)
	if len(payload.Items) > 0 {
		for k := range payload.Items[0] {
			itemKeySet[fmt.Sprintf("{item.%s}", k)] = true
			itemKeySet[fmt.Sprintf("{%s}", k)] = true
		}
	}

	templateRowIdx := -1
	itemFieldColumns := make(map[string]int)
	for rIdx, row := range rows {
		rowContainsItem := false
		for _, cellValue := range row {
			if strings.Contains(cellValue, "{item.") || itemKeySet[cellValue] {
				rowContainsItem = true
				break
			}
		}
		if rowContainsItem {
			templateRowIdx = rIdx + 1 // Excel menggunakan indeks 1-based
			for cIdx, cellValue := range row {
				for _, field := range payload.MergeDuplicateItemFields {
					if cellValue == fmt.Sprintf("{item.%s}", field) || cellValue == fmt.Sprintf("{%s}", field) {
						itemFieldColumns[field] = cIdx + 1
					}
				}
			}
			break
		}
	}

	// 2. Clone Baris Tabel (jika ditemukan)
	if templateRowIdx != -1 {
		// Some corporate templates visually group their example rows with
		// vertical merges (for example Aspek B10:B14 and Detail C10:C14).
		// Those merges are incompatible with dynamic result rows: Excel only
		// displays the value in the first cell and hides the following related
		// Aspek/Detail/Uraian values. A flat export explicitly removes vertical
		// merges from the item area before the template row is duplicated.
		if payload.FlattenItemRows {
			if err := unmergeVerticalItemRanges(f, sheetName, templateRowIdx); err != nil {
				return nil, fmt.Errorf("gagal memisahkan baris relasi template: %w", err)
			}
		}
		if payload.TrimReservedItemRows {
			if err := trimReservedItemRows(f, sheetName, templateRowIdx, rows, payload.Headers); err != nil {
				return nil, fmt.Errorf("gagal menghapus baris kosong template: %w", err)
			}
		}

		itemsCount := len(payload.Items)
		if itemsCount > 0 {
			// DuplicateRow akan menyisipkan baris berformat persis ke bawahnya, mendorong footer turun.
			for i := 1; i < itemsCount; i++ {
				if err := f.DuplicateRow(sheetName, templateRowIdx); err != nil {
					return nil, fmt.Errorf("gagal kloning baris template: %w", err)
				}
			}
		} else {
			// Jika tidak ada data item sama sekali, hapus baris template agar rapi
			f.RemoveRow(sheetName, templateRowIdx)
		}
	}

	// 3. Ambil ulang seluruh baris setelah proses duplikasi
	rows, _ = f.GetRows(sheetName)

	// 4. Proses Search & Replace Data Header dan Data Tabel
	for rIdx, row := range rows {
		excelRow := rIdx + 1
		// Cek apakah baris ini adalah rentang baris tabel item yang harus di-inject
		isItemRow := templateRowIdx != -1 && len(payload.Items) > 0 &&
			excelRow >= templateRowIdx && excelRow < templateRowIdx+len(payload.Items)

		itemIndex := excelRow - templateRowIdx

		for cIdx, cellValue := range row {
			if cellValue == "" {
				continue // Abaikan sel kosong
			}

			excelCol, _ := excelize.ColumnNumberToName(cIdx + 1)
			cellAxis := excelCol + strconv.Itoa(excelRow)
			newVal := cellValue
			matchedExact := false
			var exactValue interface{}

			// A. Replace Placeholder Item Tabel – support both {item.key} and {key} format
			if isItemRow {
				for k, v := range payload.Items[itemIndex] {
					placeholderWithItem := fmt.Sprintf("{item.%s}", k)
					placeholderDirect := fmt.Sprintf("{%s}", k)

					matched := false
					var matchedPlaceholder string
					if cellValue == placeholderWithItem || strings.Contains(newVal, placeholderWithItem) {
						matchedPlaceholder = placeholderWithItem
						matched = true
					} else if cellValue == placeholderDirect || strings.Contains(newVal, placeholderDirect) {
						matchedPlaceholder = placeholderDirect
						matched = true
					}

					if matched {
						imagePaths := getImagePaths(v)
						strVal := fmt.Sprintf("%v", v)
						if cellValue == matchedPlaceholder {
							if len(imagePaths) > 0 {
								insertImages(f, sheetName, cellAxis, imagePaths, excelRow)
								newVal = ""
							} else {
								matchedExact = true
								exactValue = v
							}
							break
						} else if strings.Contains(newVal, matchedPlaceholder) {
							if len(imagePaths) > 0 {
								insertImages(f, sheetName, cellAxis, imagePaths, excelRow)
								newVal = strings.ReplaceAll(newVal, matchedPlaceholder, "")
							} else {
								newVal = strings.ReplaceAll(newVal, matchedPlaceholder, strVal)
							}
						}
					}
				}
			}

			// B. Replace Placeholder Header (selain {item.})
			if !matchedExact {
				for k, v := range payload.Headers {
					placeholder := fmt.Sprintf("{%s}", k)
					if cellValue == placeholder {
						matchedExact = true
						exactValue = v
						break
					} else if strings.Contains(newVal, placeholder) {
						newVal = strings.ReplaceAll(newVal, placeholder, fmt.Sprintf("%v", v))
					}
				}
			}

			// 5. Tulis Hasil Replace ke Excel
			if matchedExact {
				f.SetCellValue(sheetName, cellAxis, exactValue)
			} else if newVal != cellValue {
				f.SetCellValue(sheetName, cellAxis, newVal)
			}
		}
	}

	if templateRowIdx != -1 && len(payload.Items) > 1 && len(itemFieldColumns) > 0 {
		if err := mergeDuplicateItemValues(f, sheetName, templateRowIdx, payload, itemFieldColumns); err != nil {
			return nil, fmt.Errorf("gagal menggabungkan nilai item duplikat: %w", err)
		}
	}
	if payload.TotalRow != nil && templateRowIdx != -1 {
		if err := appendTableTotalRow(f, sheetName, templateRowIdx+len(payload.Items), payload.TotalRow); err != nil {
			return nil, fmt.Errorf("gagal menambahkan total keseluruhan: %w", err)
		}
	}
	for cellAxis, value := range payload.FixedCells {
		if err := f.SetCellValue(sheetName, cellAxis, value); err != nil {
			return nil, fmt.Errorf("gagal mengisi sel tetap %s: %w", cellAxis, err)
		}
	}
	if payload.LogoCell != "" && len(payload.LogoPaths) > 0 {
		insertFirstAvailableImage(f, sheetName, payload.LogoCell, payload.LogoPaths)
	}

	// Return sebagai Buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("gagal merubah ke buffer: %w", err)
	}

	return buf, nil
}

func appendTableTotalRow(f *excelize.File, sheetName string, row int, total *TableTotalRow) error {
	if row < 1 {
		return fmt.Errorf("baris total tidak valid")
	}
	startColumn := strings.ToUpper(strings.TrimSpace(total.LabelStartColumn))
	endColumn := strings.ToUpper(strings.TrimSpace(total.LabelEndColumn))
	valueColumn := strings.ToUpper(strings.TrimSpace(total.ValueColumn))
	if startColumn == "" || endColumn == "" || valueColumn == "" {
		return fmt.Errorf("konfigurasi kolom total tidak lengkap")
	}
	if err := f.InsertRows(sheetName, row, 1); err != nil {
		return err
	}

	labelStart := fmt.Sprintf("%s%d", startColumn, row)
	labelEnd := fmt.Sprintf("%s%d", endColumn, row)
	valueCell := fmt.Sprintf("%s%d", valueColumn, row)
	if labelStart != labelEnd {
		if err := f.MergeCell(sheetName, labelStart, labelEnd); err != nil {
			return err
		}
	}
	label := strings.TrimSpace(total.Label)
	if label == "" {
		label = "TOTAL KESELURUHAN"
	}
	if err := f.SetCellValue(sheetName, labelStart, label); err != nil {
		return err
	}
	if err := f.SetCellValue(sheetName, valueCell, total.Value); err != nil {
		return err
	}

	labelStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Border:    tableBorders("64748B"),
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	if err != nil {
		return err
	}
	valueStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"0F766E"}, Pattern: 1},
		Border:    tableBorders("64748B"),
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(sheetName, labelStart, labelEnd, labelStyle); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheetName, valueCell, valueCell, valueStyle); err != nil {
		return err
	}
	return f.SetRowHeight(sheetName, row, 24)
}

func trimReservedItemRows(
	f *excelize.File,
	sheetName string,
	templateRow int,
	rows [][]string,
	headers map[string]interface{},
) error {
	footerRow := -1
	for rowIndex := templateRow; rowIndex < len(rows); rowIndex++ {
		for _, cellValue := range rows[rowIndex] {
			for key := range headers {
				if strings.Contains(cellValue, fmt.Sprintf("{%s}", key)) {
					footerRow = rowIndex + 1
					break
				}
			}
			if footerRow != -1 {
				break
			}
		}
		if footerRow != -1 {
			break
		}
	}
	if footerRow <= templateRow+1 {
		return nil
	}

	rowsToRemove := footerRow - templateRow - 1
	for index := 0; index < rowsToRemove; index++ {
		if err := f.RemoveRow(sheetName, templateRow+1); err != nil {
			return err
		}
	}
	return nil
}

func mergeDuplicateItemValues(
	f *excelize.File,
	sheetName string,
	templateRow int,
	payload *PlaceholderPayload,
	fieldColumns map[string]int,
) error {
	for _, field := range payload.MergeDuplicateItemFields {
		columnNumber, exists := fieldColumns[field]
		if !exists {
			continue
		}

		groupStart := 0
		for itemIndex := 1; itemIndex <= len(payload.Items); itemIndex++ {
			continuesGroup := itemIndex < len(payload.Items) &&
				normalizedItemValue(payload.Items[itemIndex][field]) == normalizedItemValue(payload.Items[groupStart][field]) &&
				itemFieldsEqual(payload.Items[itemIndex], payload.Items[groupStart], payload.MergeDuplicateWithinFields)
			if continuesGroup {
				continue
			}

			if itemIndex-groupStart > 1 {
				columnName, err := excelize.ColumnNumberToName(columnNumber)
				if err != nil {
					return err
				}
				startAxis := fmt.Sprintf("%s%d", columnName, templateRow+groupStart)
				endAxis := fmt.Sprintf("%s%d", columnName, templateRow+itemIndex-1)
				if err := f.MergeCell(sheetName, startAxis, endAxis); err != nil {
					return err
				}
			}
			groupStart = itemIndex
		}
	}
	return nil
}

func itemFieldsEqual(left, right map[string]interface{}, fields []string) bool {
	for _, field := range fields {
		if normalizedItemValue(left[field]) != normalizedItemValue(right[field]) {
			return false
		}
	}
	return true
}

func normalizedItemValue(value interface{}) string {
	return strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
}

// unmergeVerticalItemRanges preserves horizontal title/footer merges while
// separating every vertically merged cell in the dynamic item area. This is
// opt-in because some non-tabular reports may intentionally retain merges
// below their item placeholder.
func unmergeVerticalItemRanges(f *excelize.File, sheetName string, templateRow int) error {
	mergedCells, err := f.GetMergeCells(sheetName)
	if err != nil {
		return err
	}

	for _, mergedCell := range mergedCells {
		startAxis := mergedCell.GetStartAxis()
		endAxis := mergedCell.GetEndAxis()
		_, startRow, startErr := excelize.CellNameToCoordinates(startAxis)
		_, endRow, endErr := excelize.CellNameToCoordinates(endAxis)
		if startErr != nil || endErr != nil {
			continue
		}
		if endRow <= startRow || endRow < templateRow {
			continue
		}
		if err := f.UnmergeCell(sheetName, startAxis, endAxis); err != nil {
			return err
		}
	}
	return nil
}

// getImagePaths normalizes a placeholder value so one template cell can
// contain all evidence photos belonging to an uraian.
func getImagePaths(value interface{}) []string {
	var candidates []string
	switch typed := value.(type) {
	case string:
		candidates = []string{typed}
	case []string:
		candidates = typed
	case []interface{}:
		for _, item := range typed {
			if path, ok := item.(string); ok {
				candidates = append(candidates, path)
			}
		}
	}

	paths := make([]string, 0, len(candidates))
	seen := make(map[string]bool)
	for _, candidate := range candidates {
		if isImagePath(candidate) && !seen[candidate] {
			paths = append(paths, candidate)
			seen[candidate] = true
		}
	}
	return paths
}

// Fungsi helper mendeteksi apakah suatu string adalah path gambar
func isImagePath(path string) bool {
	if path == "" {
		return false
	}
	// Strip query parameters for extension check
	cleanPath := path
	if idx := strings.Index(cleanPath, "?"); idx != -1 {
		cleanPath = cleanPath[:idx]
	}
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
		return true
	}
	// For URLs without extension (e.g. MinIO paths like /bucket/issues/ISS-001/17398...)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return true
	}
	return false
}

// insertImages stacks all evidence photos vertically in the template's image
// cell. The row grows with the number of photos, keeping other report columns
// and the user-provided Excel layout intact.
func insertImages(f *excelize.File, sheet, cellAxis string, imagePaths []string, excelRow int) {
	if len(imagePaths) == 0 {
		return
	}
	requiredHeight := float64(len(imagePaths)) * 60
	if currentHeight, err := f.GetRowHeight(sheet, excelRow); err != nil || currentHeight < requiredHeight {
		_ = f.SetRowHeight(sheet, excelRow, requiredHeight)
	}
	for index, imagePath := range imagePaths {
		insertImageAt(f, sheet, cellAxis, imagePath, index*80)
	}
}

func insertImageAt(f *excelize.File, sheet, cellAxis, imagePath string, offsetY int) {
	data, extension, ok := loadImageData(imagePath)
	if !ok {
		return
	}
	data, extension, ok = normalizeExcelImage(data, extension)
	if !ok {
		return
	}
	addPictureData(f, sheet, cellAxis, data, extension, offsetY)
}

func insertFirstAvailableImage(f *excelize.File, sheet, cellAxis string, imagePaths []string) {
	for _, imagePath := range imagePaths {
		data, extension, ok := loadImageData(imagePath)
		if !ok {
			continue
		}
		data, extension, ok = normalizeExcelImage(data, extension)
		if !ok {
			continue
		}
		addPictureData(f, sheet, cellAxis, data, extension, 0)
		return
	}
}

func addPictureData(f *excelize.File, sheet, cellAxis string, data []byte, extension string, offsetY int) {
	scale := 0.15
	if config, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil && config.Width > 0 && config.Height > 0 {
		widthScale := 105.0 / float64(config.Width)
		heightScale := 70.0 / float64(config.Height)
		scale = widthScale
		if heightScale < scale {
			scale = heightScale
		}
	}

	picFormat := &excelize.GraphicOptions{
		OffsetX:         5,
		OffsetY:         5 + offsetY,
		ScaleX:          scale,
		ScaleY:          scale,
		PrintObject:     func() *bool { b := true; return &b }(),
		Locked:          func() *bool { b := false; return &b }(),
		LockAspectRatio: true,
		Positioning:     "oneCell",
	}

	_ = f.AddPictureFromBytes(sheet, cellAxis, &excelize.Picture{
		Extension: extension,
		File:      data,
		Format:    picFormat,
	})
}

// normalizeExcelImage detects the binary format instead of trusting the file
// name. Existing MinIO objects can have a .jpg name while containing WebP
// bytes; Excel does not support WebP pictures, so they are converted to PNG.
func normalizeExcelImage(data []byte, extension string) ([]byte, string, bool) {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return data, extension, extension != ""
	}

	switch strings.ToLower(format) {
	case "webp":
		decoded, _, decodeErr := image.Decode(bytes.NewReader(data))
		if decodeErr != nil {
			return nil, "", false
		}
		var converted bytes.Buffer
		if encodeErr := png.Encode(&converted, decoded); encodeErr != nil {
			return nil, "", false
		}
		return converted.Bytes(), ".png", true
	case "jpeg":
		return data, ".jpg", true
	case "png":
		return data, ".png", true
	case "gif":
		return data, ".gif", true
	default:
		return data, extension, extension != ""
	}
}

func loadImageData(imagePath string) ([]byte, string, bool) {
	urlsToTry := imageURLsToTry(imagePath)
	if len(urlsToTry) > 0 {
		// Add host fallbacks for Docker vs local environment
		if strings.Contains(imagePath, "localhost:9000") {
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "localhost:9000", "minio:9000"))
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "localhost:9000", "127.0.0.1:9000"))
		} else if strings.Contains(imagePath, "minio:9000") {
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "minio:9000", "localhost:9000"))
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "minio:9000", "127.0.0.1:9000"))
		}

		client := &http.Client{Timeout: 15 * time.Second}
		for _, targetURL := range urlsToTry {
			resp, err := client.Get(targetURL)
			if err == nil {
				body, readErr := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode == 200 {
					if readErr == nil && len(body) > 0 {
						ext := imageExtension(targetURL, resp.Header.Get("Content-Type"))
						return body, ext, true
					}
				}
			}
		}
		return nil, "", false
	}

	localPath := imagePath
	if strings.HasPrefix(localPath, "/") {
		localPath = "." + localPath
	}
	data, err := os.ReadFile(localPath)
	if err != nil && localPath != imagePath {
		data, err = os.ReadFile(imagePath)
	}
	if err != nil || len(data) == 0 {
		return nil, "", false
	}
	return data, imageExtension(imagePath, ""), true
}

// imageURLsToTry resolves both absolute image URLs and the relative MinIO URLs
// stored in Issue_Photo (for example /monitoring-audit-bucket/issues/...).
// The browser resolves the latter through a Next.js proxy, while an Excel
// export must fetch the object directly from MinIO inside the backend service.
func imageURLsToTry(imagePath string) []string {
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		return []string{imagePath}
	}

	bucket := strings.TrimSpace(os.Getenv("MINIO_BUCKET"))
	if bucket == "" {
		bucket = "monitoring-audit-bucket"
	}
	bucketPath := "/" + strings.Trim(bucket, "/") + "/"
	pathIndex := strings.Index(imagePath, bucketPath)
	if pathIndex == -1 {
		return nil
	}
	objectPath := imagePath[pathIndex:]

	endpoint := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	baseURL := strings.TrimRight(endpoint, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		scheme := "http://"
		if strings.EqualFold(strings.TrimSpace(os.Getenv("MINIO_USE_SSL")), "true") {
			scheme = "https://"
		}
		baseURL = scheme + baseURL
	}

	urls := []string{baseURL + objectPath}
	if strings.Contains(baseURL, "minio:9000") {
		urls = append(urls, strings.Replace(baseURL, "minio:9000", "localhost:9000", 1)+objectPath)
	} else if strings.Contains(baseURL, "localhost:9000") {
		urls = append(urls, strings.Replace(baseURL, "localhost:9000", "minio:9000", 1)+objectPath)
	}
	return urls
}

func imageExtension(path, contentType string) string {
	cleanPath := path
	if idx := strings.Index(cleanPath, "?"); idx != -1 {
		cleanPath = cleanPath[:idx]
	}
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if ext != "" {
		return ext
	}
	switch {
	case strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	default:
		return ".png"
	}
}

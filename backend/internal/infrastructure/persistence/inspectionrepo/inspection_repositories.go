package inspectionrepo

import (
	"sync"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type inspHeaderCacheItem struct {
	header    *inspection.InspectionHeader
	expiresAt time.Time
}

var globalInspHeaderCache  sync.Map
var inspHeaderSFGroup      singleflight.Group // mencegah cache stampede saat VU concurrent miss

// ── Inspection Header ─────────────────────────────────────────────────────

type inspectionHeaderRepository struct{ db *gorm.DB }

func NewInspectionHeaderRepository(db *gorm.DB) inspection.InspectionHeaderRepository {
	return &inspectionHeaderRepository{db: db}
}

func (r *inspectionHeaderRepository) FindAll(page, limit int, plantID, areaID, status, inspectorID string) ([]inspection.InspectionHeader, int64, error) {
	var items []inspection.InspectionHeader
	var total int64
	q := r.db.Model(&inspection.InspectionHeader{})
	if plantID != "" {
		q = q.Where(`"AreaID" IN (SELECT "AreaID" FROM "Area_Master" WHERE "PlantID" = ? OR "PlantID" IS NULL OR "PlantID" = '')`, plantID)
	}
	if areaID != "" {
		q = q.Where("\"AreaID\" = ?", areaID)
	}
	if status != "" {
		q = q.Where("\"InspectionHeaderStatus\" = ?", status)
	}
	if inspectorID != "" {
		q = q.Where("\"InspectorID\" = ?", inspectorID)
	}
	q.Count(&total)
	err := q.Order(`"InspectionHeaderCreatedAt" DESC`).Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *inspectionHeaderRepository) FindByID(id string) (*inspection.InspectionHeader, error) {
	// Fast path: return from in-memory cache
	if val, ok := globalInspHeaderCache.Load(id); ok {
		item := val.(inspHeaderCacheItem)
		if time.Now().Before(item.expiresAt) {
			return item.header, nil
		}
		globalInspHeaderCache.Delete(id)
	}

	// singleflight: hanya 1 goroutine yang query DB, VU lain menunggu hasil yang sama.
	// Ini mencegah 200 VU concurrent semua hit DB untuk inspection ID yang sama.
	result, err, _ := inspHeaderSFGroup.Do(id, func() (interface{}, error) {
		// Re-check cache di dalam singleflight untuk menghindari race
		if val, ok := globalInspHeaderCache.Load(id); ok {
			item := val.(inspHeaderCacheItem)
			if time.Now().Before(item.expiresAt) {
				return item.header, nil
			}
		}

		var item inspection.InspectionHeader
		dbErr := r.db.Model(&item).
			Select(`"Inspection_Header".*, 
				"Area_Master"."AreaName" AS "AreaName", 
				"Kawasan_Master"."KawasanName" AS "KawasanName", 
				"DetailKawasan_Master"."DetailKawasanName" AS "DetailKawasanName", 
				"Users"."FullName" AS "InspectorName"`).
			Joins(`LEFT JOIN "Area_Master" ON "Inspection_Header"."AreaID" = "Area_Master"."AreaID"`).
			Joins(`LEFT JOIN "Kawasan_Master" ON "Inspection_Header"."KawasanID" = "Kawasan_Master"."KawasanID"`).
			Joins(`LEFT JOIN "DetailKawasan_Master" ON "Inspection_Header"."DetailKawasanID" = "DetailKawasan_Master"."DetailKawasanID"`).
			Joins(`LEFT JOIN "Users" ON "Inspection_Header"."InspectorID" = "Users"."UserID"`).
			Where(`"Inspection_Header"."InspectionID" = ?`, id).
			Take(&item).Error

		if dbErr == nil {
			type scoreRow struct {
				Total   int
				OkCount int
			}
			var sr scoreRow
			r.db.Raw(`
				SELECT
					COUNT(*) FILTER (WHERE "Checking" IN ('OK','NG')) AS total,
					COUNT(*) FILTER (WHERE "Checking" = 'OK') AS ok_count
				FROM "Inspection_Result"
				WHERE "InspectionID" = ?`, id).Scan(&sr)
			if sr.Total > 0 {
				scoreVal := float64(sr.OkCount) * 100.0 / float64(sr.Total)
				item.Score = &scoreVal
			}
			// Cache 5 menit
			globalInspHeaderCache.Store(id, inspHeaderCacheItem{
				header:    &item,
				expiresAt: time.Now().Add(5 * time.Minute),
			})
			return &item, nil
		}
		return nil, dbErr
	})

	if err != nil {
		return nil, err
	}
	return result.(*inspection.InspectionHeader), nil
}

func (r *inspectionHeaderRepository) FindActiveByKawasan(kawasanID string) ([]inspection.InspectionHeader, error) {
	var items []inspection.InspectionHeader
	err := r.db.Where("\"KawasanID\" = ? AND \"InspectionHeaderStatus\" IN (?, ?)",
		kawasanID, inspection.InspectionStatusDraft, inspection.InspectionStatusOngoing).
		Find(&items).Error
	return items, err
}

func (r *inspectionHeaderRepository) FindActiveByDetailKawasan(detailKawasanID string) ([]inspection.InspectionHeader, error) {
	var items []inspection.InspectionHeader
	err := r.db.Where("\"DetailKawasanID\" = ? AND \"InspectionHeaderStatus\" IN (?, ?)",
		detailKawasanID, inspection.InspectionStatusDraft, inspection.InspectionStatusOngoing).
		Find(&items).Error
	return items, err
}

func (r *inspectionHeaderRepository) FindActiveByInspector(inspectorID string) ([]inspection.InspectionHeader, error) {
	var items []inspection.InspectionHeader
	err := r.db.Where("\"InspectorID\" = ? AND \"InspectionHeaderStatus\" IN (?, ?)",
		inspectorID, inspection.InspectionStatusDraft, inspection.InspectionStatusOngoing).
		Find(&items).Error
	return items, err
}

func (r *inspectionHeaderRepository) FindActiveByInspectorAndDetailKawasan(inspectorID, detailKawasanID string) ([]inspection.InspectionHeader, error) {
	var items []inspection.InspectionHeader
	err := r.db.Where("\"InspectorID\" = ? AND \"DetailKawasanID\" = ? AND \"InspectionHeaderStatus\" IN (?, ?)",
		inspectorID, detailKawasanID, inspection.InspectionStatusDraft, inspection.InspectionStatusOngoing).
		Find(&items).Error
	return items, err
}

func (r *inspectionHeaderRepository) CountCompletedThisMonthByDetailKawasan(detailKawasanID string, year int, month int) (int64, error) {
	var count int64
	err := r.db.Model(&inspection.InspectionHeader{}).
		Where("\"DetailKawasanID\" = ? AND \"InspectionHeaderStatus\" IN (?, ?)", detailKawasanID, inspection.InspectionStatusCompleted, inspection.InspectionStatusApproved).
		Where("EXTRACT(YEAR FROM \"InspectionHeaderCreatedAt\") = ? AND EXTRACT(MONTH FROM \"InspectionHeaderCreatedAt\") = ?", year, month).
		Count(&count).Error
	return count, err
}

func (r *inspectionHeaderRepository) CountCompletedByAreaAndDetailKawasan(areaID, detailKawasanID string) (int64, error) {
	var count int64
	err := r.db.Model(&inspection.InspectionHeader{}).
		Where("\"AreaID\" = ? AND \"DetailKawasanID\" = ? AND \"InspectionHeaderStatus\" = ?",
			areaID, detailKawasanID, inspection.InspectionStatusCompleted).
		Count(&count).Error
	return count, err
}

// GetTrendByContext aggregates inspection count by month for a given year and inspector.
func (r *inspectionHeaderRepository) GetTrendByContext(contextID string, year int) ([]inspection.TrendData, error) {
	var results []inspection.TrendData

	// Default empty array for 12 months
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	trendMap := make(map[string]int)
	for _, m := range monthNames {
		trendMap[m] = 0
	}

	type QueryResult struct {
		Month int
		Count int
	}
	var qr []QueryResult

	// Postgres specific or SQLite. Assuming Postgres since it's the usual for this boilerplate, but let's use standard.
	// We'll use GORM extract month. If sqlite, EXTRACT doesn't work out of the box like this.
	// Assuming Postgres for EXTRACT(MONTH FROM ...). If SQLite, strftime('%m', ...)
	// Let's use a safe cross-database approach by querying records for the year and grouping in code if needed,
	// but standard SQL `EXTRACT(MONTH FROM InspectionHeaderCreatedAt)` is common.
	err := r.db.Model(&inspection.InspectionHeader{}).
		Select(`EXTRACT(MONTH FROM "InspectionHeaderCreatedAt") as month, COUNT(*) as count`).
		Where("\"InspectorID\" = ? AND EXTRACT(YEAR FROM \"InspectionHeaderCreatedAt\") = ?", contextID, year).
		Group(`EXTRACT(MONTH FROM "InspectionHeaderCreatedAt")`).
		Scan(&qr).Error

	if err != nil {
		// Fallback for SQLite which doesn't support EXTRACT natively like this
		// This is a safety net since we don't know the exact DB driver being used.
		var allHeaders []inspection.InspectionHeader
		err = r.db.Where("\"InspectorID\" = ?", contextID).Find(&allHeaders).Error
		if err != nil {
			return results, err
		}
		for _, h := range allHeaders {
			if h.InspectionHeaderCreatedAt.Year() == year {
				m := int(h.InspectionHeaderCreatedAt.Month())
				trendMap[monthNames[m-1]]++
			}
		}
	} else {
		for _, row := range qr {
			if row.Month >= 1 && row.Month <= 12 {
				trendMap[monthNames[row.Month-1]] = row.Count
			}
		}
	}

	for _, m := range monthNames {
		results = append(results, inspection.TrendData{
			Month: m,
			Rate:  trendMap[m],
		})
	}

	return results, nil
}

func (r *inspectionHeaderRepository) Create(h *inspection.InspectionHeader) error {
	return r.db.Create(h).Error
}
func (r *inspectionHeaderRepository) Update(h *inspection.InspectionHeader) error {
	globalInspHeaderCache.Delete(h.InspectionID)
	return r.db.Save(h).Error
}
func (r *inspectionHeaderRepository) Delete(id string) error {
	globalInspHeaderCache.Delete(id)
	var h inspection.InspectionHeader
	if err := r.db.Where("\"InspectionID\" = ?", id).First(&h).Error; err == nil {
		// 1. Delete associated Issue_Photo records
		_ = r.db.Exec(`
			DELETE FROM "Issue_Photo" 
			WHERE "IssueID" IN (
				SELECT i."IssueID" FROM "Issue" i
				JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"
				WHERE ir."InspectionID" = ?
			)`, id)

		// 2. Delete associated Issue_HEI records
		_ = r.db.Exec(`
			DELETE FROM "Issue_HEI" 
			WHERE "IssueID" IN (
				SELECT i."IssueID" FROM "Issue" i
				JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"
				WHERE ir."InspectionID" = ?
			)`, id)

		// 3. Delete associated Issue records
		_ = r.db.Exec(`
			DELETE FROM "Issue" 
			WHERE "ResultID" IN (
				SELECT "ResultID" FROM "Inspection_Result" WHERE "InspectionID" = ?
			)`, id)

		// 4. Delete associated Inspection_Result records
		_ = r.db.Where("\"InspectionID\" = ?", id).Delete(&inspection.InspectionResult{})

		// 5. Delete Inspection_Header
		if err := r.db.Where("\"InspectionID\" = ?", id).Delete(&inspection.InspectionHeader{}).Error; err != nil {
			return err
		}

		// Sync / recalculate LastInspection on DetailKawasan_Master
		_ = r.db.Exec(`
			UPDATE "DetailKawasan_Master" 
			SET "LastInspection" = (
				SELECT MAX("InspectionHeaderCreatedAt") 
				FROM "Inspection_Header" 
				WHERE "DetailKawasanID" = ? AND "InspectionHeaderStatus" IN ('Completed', 'Approved')
			) 
			WHERE "DetailKawasanID" = ?`, h.DetailKawasanID, h.DetailKawasanID)

		// Sync / recalculate LastInspection on Kawasan_Master
		_ = r.db.Exec(`
			UPDATE "Kawasan_Master" 
			SET "LastInspection" = (
				SELECT MAX("InspectionHeaderCreatedAt") 
				FROM "Inspection_Header" 
				WHERE "KawasanID" = ? AND "InspectionHeaderStatus" IN ('Completed', 'Approved')
			) 
			WHERE "KawasanID" = ?`, h.KawasanID, h.KawasanID)

		return nil
	}
	return r.db.Where("\"InspectionID\" = ?", id).Delete(&inspection.InspectionHeader{}).Error
}

// ── Inspection Result ─────────────────────────────────────────────────────

type inspectionResultRepository struct{ db *gorm.DB }

func NewInspectionResultRepository(db *gorm.DB) inspection.InspectionResultRepository {
	return &inspectionResultRepository{db: db}
}

func (r *inspectionResultRepository) FindByInspectionID(inspectionID string) ([]inspection.InspectionResult, error) {
	var items []inspection.InspectionResult
	err := r.db.Where("\"InspectionID\" = ?", inspectionID).Find(&items).Error
	return items, err
}

func (r *inspectionResultRepository) FindByID(id string) (*inspection.InspectionResult, error) {
	var item inspection.InspectionResult
	err := r.db.Where("\"ResultID\" = ?", id).First(&item).Error
	return &item, err
}

func (r *inspectionResultRepository) Create(res *inspection.InspectionResult) error {
	return r.db.Create(res).Error
}
func (r *inspectionResultRepository) BulkCreate(rs []inspection.InspectionResult) error {
	if len(rs) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&rs, 50).Error
}
func (r *inspectionResultRepository) Update(res *inspection.InspectionResult) error {
	return r.db.Save(res).Error
}
func (r *inspectionResultRepository) Delete(id string) error {
	return r.db.Where("\"ResultID\" = ?", id).Delete(&inspection.InspectionResult{}).Error
}

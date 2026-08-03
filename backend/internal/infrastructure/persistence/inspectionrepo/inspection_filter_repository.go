package inspectionrepo

import (
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"gorm.io/gorm"
)

// ── Inspection Filter Repository ──────────────────────────────────────────

type inspectionFilterRepository struct{ db *gorm.DB }

// NewInspectionFilterRepository creates a new InspectionFilterRepository.
func NewInspectionFilterRepository(db *gorm.DB) inspection.InspectionFilterRepository {
	return &inspectionFilterRepository{db: db}
}

// FindFiltered returns a paginated list of InspectionHeaders matching the filter.
func (r *inspectionFilterRepository) FindFiltered(f *inspection.InspectionFilter) ([]inspection.InspectionHeader, int64, error) {
	var items []inspection.InspectionHeader
	var total int64

	base := r.db.Model(&inspection.InspectionHeader{})
	base = f.ApplyTo(base)
	base.Count(&total)

	err := f.ApplySort(base).
		Select(`"Inspection_Header".*, 
			"Area_Master"."AreaName" AS "AreaName", 
			"Kawasan_Master"."KawasanName" AS "KawasanName", 
			"DetailKawasan_Master"."DetailKawasanName" AS "DetailKawasanName", 
			"Users"."FullName" AS "InspectorName",
			(SELECT CASE WHEN COUNT(*) > 0 THEN (COUNT(CASE WHEN "Checking" = 'OK' THEN 1 END) * 100.0 / COUNT(*)) ELSE 0.0 END FROM "Inspection_Result" ir WHERE ir."InspectionID" = "Inspection_Header"."InspectionID" AND ir."Checking" IN ('OK', 'NG')) AS "Score"`).
		Joins(`LEFT JOIN "Area_Master" ON "Inspection_Header"."AreaID" = "Area_Master"."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" ON "Inspection_Header"."KawasanID" = "Kawasan_Master"."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" ON "Inspection_Header"."DetailKawasanID" = "DetailKawasan_Master"."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" ON "Inspection_Header"."InspectorID" = "Users"."UserID"`).
		Offset((f.Page - 1) * f.Limit).
		Limit(f.Limit).
		Find(&items).Error

	return items, total, err
}

// FindFacets computes per-field counts. For each facet field, the filter for
// that field itself is excluded (facet exclusion) so the chip counts remain
// accurate when a value is already selected.
func (r *inspectionFilterRepository) FindFacets(f *inspection.InspectionFilter) (inspection.InspectionFacets, error) {
	facets := inspection.InspectionFacets{
		Status:      make(map[string]int64),
		AreaID:      make(map[string]int64),
		KawasanID:   make(map[string]int64),
		InspectorID: make(map[string]int64),
		DateRange: inspection.DateRangeFacet{
			Field: "created_at",
		},
	}

	type kv struct {
		Key   string
		Count int64
	}

	// Helper: base query without a specific field so we get all counts for that field.
	baseWithout := func(excludeStatus, excludeArea, excludeKawasan, excludeInspector, excludeDate bool) *gorm.DB {
		tmp := &inspection.InspectionFilter{
			Q:                f.Q,
			ScopeInspectorID: f.ScopeInspectorID,
			PlantID:          f.PlantID,
			DetailKawasanID:  f.DetailKawasanID,
		}
		if !excludeStatus {
			tmp.Status = f.Status
			tmp.StatusIn = f.StatusIn
		}
		if !excludeArea {
			tmp.AreaID = f.AreaID
			tmp.AreaIDIn = f.AreaIDIn
		}
		if !excludeKawasan {
			tmp.KawasanID = f.KawasanID
			tmp.KawasanIDIn = f.KawasanIDIn
		}
		if !excludeInspector {
			tmp.InspectorID = f.InspectorID
			tmp.InspectorIDIn = f.InspectorIDIn
		}
		if !excludeDate {
			tmp.DateFrom = f.DateFrom
			tmp.DateTo = f.DateTo
		}
		return tmp.ApplyTo(r.db.Model(&inspection.InspectionHeader{}))
	}

	// Status facet (without status filter)
	var statusRows []kv
	if err := baseWithout(true, false, false, false, false).
		Select(`"InspectionHeaderStatus" AS key, COUNT(*) AS count`).
		Group(`"InspectionHeaderStatus"`).
		Scan(&statusRows).Error; err == nil {
		for _, row := range statusRows {
			facets.Status[row.Key] = row.Count
		}
	}

	// AreaID facet (without area filter)
	var areaRows []kv
	if err := baseWithout(false, true, false, false, false).
		Select(`"AreaID" AS key, COUNT(*) AS count`).
		Group(`"AreaID"`).
		Scan(&areaRows).Error; err == nil {
		for _, row := range areaRows {
			facets.AreaID[row.Key] = row.Count
		}
	}

	// KawasanID facet (without kawasan filter)
	var kawasanRows []kv
	if err := baseWithout(false, false, true, false, false).
		Select(`"KawasanID" AS key, COUNT(*) AS count`).
		Group(`"KawasanID"`).
		Scan(&kawasanRows).Error; err == nil {
		for _, row := range kawasanRows {
			facets.KawasanID[row.Key] = row.Count
		}
	}

	// InspectorID facet (without inspector filter)
	var inspectorRows []kv
	if err := baseWithout(false, false, false, true, false).
		Select(`"InspectorID" AS key, COUNT(*) AS count`).
		Group(`"InspectorID"`).
		Scan(&inspectorRows).Error; err == nil {
		for _, row := range inspectorRows {
			facets.InspectorID[row.Key] = row.Count
		}
	}

	// Date range (without date filter)
	type dateResult struct {
		Min *time.Time
		Max *time.Time
	}
	var dr dateResult
	if err := baseWithout(false, false, false, false, true).
		Select(`MIN("InspectionHeaderCreatedAt") AS min, MAX("InspectionHeaderCreatedAt") AS max`).
		Scan(&dr).Error; err == nil {
		facets.DateRange.Min = dr.Min
		facets.DateRange.Max = dr.Max
	}

	return facets, nil
}

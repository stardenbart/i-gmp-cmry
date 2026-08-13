package inspectionrepo

import (
	"sync"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
)

type aspekChecklistCacheItem struct {
	aspeks    []master.Aspek
	expiresAt time.Time
}

// Cache layer 2: full checklist per inspectionID (termasuk results)
type fullChecklistCacheItem struct {
	checklist *inspection.FullChecklist
	expiresAt time.Time
}

var globalAspekChecklistCache    sync.Map
var globalFullChecklistCache     sync.Map

func (r *inspectionHeaderRepository) GetFullChecklist(areaID, inspectionID string) (*inspection.FullChecklist, error) {
	// Cache layer 2: full checklist per inspectionID (2 menit TTL)
	// Ini mencegah 200 user concurrent semua hit DB untuk inspection yang sama
	if inspectionID != "" {
		if val, ok := globalFullChecklistCache.Load(inspectionID); ok {
			item := val.(fullChecklistCacheItem)
			if time.Now().Before(item.expiresAt) {
				return item.checklist, nil
			}
			globalFullChecklistCache.Delete(inspectionID)
		}
	}

	// Cache layer 1: aspek structure per areaID (1 jam TTL — master data jarang berubah)
	var aspeks []master.Aspek
	if val, ok := globalAspekChecklistCache.Load(areaID); ok {
		item := val.(aspekChecklistCacheItem)
		if time.Now().Before(item.expiresAt) {
			aspeks = item.aspeks
		} else {
			globalAspekChecklistCache.Delete(areaID)
		}
	}

	if len(aspeks) == 0 {
		err := r.db.Preload("Details.Urains").Where("\"AreaID\" = ?", areaID).Find(&aspeks).Error
		if err != nil {
			return nil, err
		}
		globalAspekChecklistCache.Store(areaID, aspekChecklistCacheItem{
			aspeks:    aspeks,
			expiresAt: time.Now().Add(1 * time.Hour), // 10 menit → 1 jam
		})
	}

	// Fetch all InspectionResults untuk inspectionID ini (1 query ringan)
	var results []inspection.InspectionResult
	if inspectionID != "" {
		r.db.Where("\"InspectionID\" = ?", inspectionID).Find(&results)
	}
	resultMap := make(map[string]*inspection.InspectionResult)
	for i := range results {
		resultMap[results[i].UraianID] = &results[i]
	}

	// Map ke DTO
	checklist := &inspection.FullChecklist{
		InspectionID: inspectionID,
		AreaID:       areaID,
		Aspeks:       make([]inspection.ChecklistAspek, 0),
	}

	for _, a := range aspeks {
		ca := inspection.ChecklistAspek{
			AspekID:   a.AspekID,
			AspekName: a.AspekName,
			Details:   make([]inspection.ChecklistDetail, 0),
		}
		for _, d := range a.Details {
			cd := inspection.ChecklistDetail{
				DetailID:   d.DetailID,
				DetailName: d.DetailName,
				Uraians:    make([]inspection.ChecklistUraian, 0),
			}
			for _, u := range d.Urains {
				cu := inspection.ChecklistUraian{
					UraianID:      u.UraianID,
					UraianText:    u.UraianText,
					StandardScore: u.StandardScore,
					Result:        resultMap[u.UraianID],
				}
				cd.Uraians = append(cd.Uraians, cu)
			}
			ca.Details = append(ca.Details, cd)
		}
		checklist.Aspeks = append(checklist.Aspeks, ca)
	}

	// Simpan full checklist ke cache per inspectionID (2 menit)
	if inspectionID != "" {
		globalFullChecklistCache.Store(inspectionID, fullChecklistCacheItem{
			checklist: checklist,
			expiresAt: time.Now().Add(2 * time.Minute),
		})
	}

	return checklist, nil
}

// InvalidateChecklistCache dipanggil saat ada update pada inspection result
// agar cache tidak stale setelah auditor submit jawaban
func InvalidateChecklistCache(inspectionID string) {
	globalFullChecklistCache.Delete(inspectionID)
}

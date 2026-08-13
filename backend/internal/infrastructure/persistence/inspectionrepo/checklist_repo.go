package inspectionrepo

import (
	"sync"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
	"golang.org/x/sync/singleflight"
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

var globalAspekChecklistCache sync.Map
var globalFullChecklistCache  sync.Map

// singleflight groups — mencegah cache stampede saat 200 VU concurrent hit cache miss bersamaan.
// Hanya satu goroutine yang query DB, sisanya menunggu dan menerima hasil yang sama.
var checklistSFGroup   singleflight.Group
var aspekStructSFGroup singleflight.Group

func (r *inspectionHeaderRepository) GetFullChecklist(areaID, inspectionID string) (*inspection.FullChecklist, error) {
	// Cache layer 2: full checklist per inspectionID (2 menit TTL)
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
	// Gunakan singleflight agar saat 200 VU concurrent miss cache secara bersamaan,
	// hanya 1 goroutine yang query DB. Sisanya menunggu hasil yang sama.
	type aspekResult struct {
		aspeks []master.Aspek
		err    error
	}

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
		sfKey := "aspek-struct:" + areaID
		result, err, _ := aspekStructSFGroup.Do(sfKey, func() (interface{}, error) {
			// Re-check cache inside singleflight to avoid race between waiters
			if val, ok := globalAspekChecklistCache.Load(areaID); ok {
				item := val.(aspekChecklistCacheItem)
				if time.Now().Before(item.expiresAt) {
					return item.aspeks, nil
				}
			}

			var dbAspeks []master.Aspek
			dbErr := r.db.
				Preload("Details.Urains").
				Where(`"AreaID" = ?`, areaID).
				Find(&dbAspeks).Error
			if dbErr != nil {
				return nil, dbErr
			}
			globalAspekChecklistCache.Store(areaID, aspekChecklistCacheItem{
				aspeks:    dbAspeks,
				expiresAt: time.Now().Add(1 * time.Hour),
			})
			return dbAspeks, nil
		})
		if err != nil {
			return nil, err
		}
		aspeks = result.([]master.Aspek)
	}

	// Fetch all InspectionResults untuk inspectionID ini dalam 1 query ringan
	var results []inspection.InspectionResult
	if inspectionID != "" {
		r.db.Where(`"InspectionID" = ?`, inspectionID).Find(&results)
	}
	resultMap := make(map[string]*inspection.InspectionResult, len(results))
	for i := range results {
		resultMap[results[i].UraianID] = &results[i]
	}

	// Map ke DTO
	checklist := &inspection.FullChecklist{
		InspectionID: inspectionID,
		AreaID:       areaID,
		Aspeks:       make([]inspection.ChecklistAspek, 0, len(aspeks)),
	}

	for _, a := range aspeks {
		ca := inspection.ChecklistAspek{
			AspekID:   a.AspekID,
			AspekName: a.AspekName,
			Details:   make([]inspection.ChecklistDetail, 0, len(a.Details)),
		}
		for _, d := range a.Details {
			cd := inspection.ChecklistDetail{
				DetailID:   d.DetailID,
				DetailName: d.DetailName,
				Uraians:    make([]inspection.ChecklistUraian, 0, len(d.Urains)),
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

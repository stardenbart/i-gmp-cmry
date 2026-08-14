package inspectionrepo

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
	"golang.org/x/sync/singleflight"
)

// ─── Generic mutex-protected TTL cache ───────────────────────────────────────
// sync.Map di Go 1.26 menggunakan HashTrieMap yang panic dengan error
// "ran out of hash bits" saat concurrent Store/Delete tanpa sinkronisasi.
// Semua cache dalam package ini menggunakan pattern ini.

type ttlEntry[V any] struct {
	value     V
	expiresAt time.Time
}

type ttlCache[V any] struct {
	mu    sync.RWMutex
	items map[string]ttlEntry[V]
}

func newTTLCache[V any]() *ttlCache[V] {
	return &ttlCache[V]{items: make(map[string]ttlEntry[V])}
}

func (c *ttlCache[V]) Load(key string) (V, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		var zero V
		return zero, false
	}
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		if entry, ok := c.items[key]; ok && time.Now().After(entry.expiresAt) {
			delete(c.items, key)
		}
		c.mu.Unlock()
		var zero V
		return zero, false
	}
	return entry.value, true
}

func (c *ttlCache[V]) Store(key string, value V, ttl time.Duration) {
	c.mu.Lock()
	c.items[key] = ttlEntry[V]{value: value, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

func (c *ttlCache[V]) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// ─── Cache singletons ────────────────────────────────────────────────────────
var globalAspekChecklistCache = newTTLCache[[]master.Aspek]()
var globalFullChecklistCache  = newTTLCache[*inspection.FullChecklist]()

// singleflight groups — mencegah cache stampede saat 200 VU concurrent hit cache miss bersamaan.
var checklistSFGroup   singleflight.Group
var aspekStructSFGroup singleflight.Group

func (r *inspectionHeaderRepository) GetFullChecklist(areaID, inspectionID string) (*inspection.FullChecklist, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return r.GetFullChecklistWithCtx(ctx, areaID, inspectionID)
}

func (r *inspectionHeaderRepository) GetFullChecklistWithCtx(parentCtx context.Context, areaID, inspectionID string) (*inspection.FullChecklist, error) {
	// Fast path: Cache layer 2 (full checklist per inspectionID, 2 min TTL)
	if inspectionID != "" {
		if cached, ok := globalFullChecklistCache.Load(inspectionID); ok {
			return cached, nil
		}
	}

	// Slow path: singleflight per inspectionID & areaID
	sfKey := "checklist:" + areaID + ":" + inspectionID
	result, err, _ := checklistSFGroup.Do(sfKey, func() (interface{}, error) {
		if inspectionID != "" {
			if cached, ok := globalFullChecklistCache.Load(inspectionID); ok {
				return cached, nil
			}
		}

		ctx := parentCtx
		if ctx == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
		}

		// Cache layer 1: aspek structure per areaID (1 jam TTL — master data jarang berubah)
		var aspeks []master.Aspek
		if cached, ok := globalAspekChecklistCache.Load(areaID); ok {
			aspeks = cached
		}

		if len(aspeks) == 0 {
			aspekKey := "aspek-struct:" + areaID
			res, dbErr, _ := aspekStructSFGroup.Do(aspekKey, func() (interface{}, error) {
				if cached, ok := globalAspekChecklistCache.Load(areaID); ok {
					return cached, nil
				}
				var dbAspeks []master.Aspek
				err := r.db.WithContext(ctx).
					Preload("Details.Urains").
					Where(`"AreaID" = ?`, areaID).
					Find(&dbAspeks).Error
				if err != nil {
					return nil, err
				}
				globalAspekChecklistCache.Store(areaID, dbAspeks, 1*time.Hour)
				return dbAspeks, nil
			})
			if dbErr != nil {
				return nil, dbErr
			}
			aspeks = res.([]master.Aspek)
		}

		// Fetch all InspectionResults untuk inspectionID ini dalam 1 query ringan
		var results []inspection.InspectionResult
		if inspectionID != "" {
			if dbErr := r.db.WithContext(ctx).Where(`"InspectionID" = ?`, inspectionID).Find(&results).Error; dbErr != nil {
				return nil, dbErr
			}
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
			globalFullChecklistCache.Store(inspectionID, checklist, 2*time.Minute)
		}

		return checklist, nil
	})

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("checklist not found")
	}
	checklist, ok := result.(*inspection.FullChecklist)
	if !ok || checklist == nil {
		return nil, errors.New("checklist not found")
	}
	return checklist, nil
}

// InvalidateChecklistCache dipanggil saat ada update pada inspection result
// agar cache tidak stale setelah auditor submit jawaban
func InvalidateChecklistCache(inspectionID string) {
	globalFullChecklistCache.Delete(inspectionID)
}



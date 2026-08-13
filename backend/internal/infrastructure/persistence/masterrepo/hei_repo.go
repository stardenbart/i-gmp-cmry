package masterrepo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

type heiCacheItem struct {
	items     []master.HEIMaster
	total     int64
	expiresAt time.Time
}

type heiCatCacheItem struct {
	categories []string
	expiresAt  time.Time
}

var (
	globalHEICache    sync.Map
	globalHEICatCache sync.Map
)

type heiRepository struct {
	db *gorm.DB
}

func NewHEIRepository(db *gorm.DB) master.HEIRepository {
	return &heiRepository{db: db}
}

func (r *heiRepository) FindAll(page, limit int, category, search string) ([]master.HEIMaster, int64, error) {
	cacheKey := fmt.Sprintf("hei:%d:%d:%s:%s", page, limit, category, search)
	if val, ok := globalHEICache.Load(cacheKey); ok {
		item := val.(heiCacheItem)
		if time.Now().Before(item.expiresAt) {
			return item.items, item.total, nil
		}
		globalHEICache.Delete(cacheKey)
	}

	var items []master.HEIMaster
	var total int64

	q := r.db.Model(&master.HEIMaster{})

	if category != "" {
		q = q.Where("LOWER(\"CategoryName\") = ?", strings.ToLower(category))
	}

	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		q = q.Where("LOWER(\"HEIName\") LIKE ? OR LOWER(\"HEICode\") LIKE ? OR LOWER(\"CategoryName\") LIKE ?", s, s, s)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if limit <= 0 {
		limit = 500
	}

	if err := q.Order("\"CategoryName\" ASC, \"HEICode\" ASC, \"CreatedAt\" DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	globalHEICache.Store(cacheKey, heiCacheItem{
		items:     items,
		total:     total,
		expiresAt: time.Now().Add(10 * time.Minute),
	})

	return items, total, nil
}

func (r *heiRepository) FindByID(id string) (*master.HEIMaster, error) {
	var h master.HEIMaster
	if err := r.db.Where("\"HEIID\" = ?", id).First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *heiRepository) FindCategories() ([]string, error) {
	cacheKey := "hei_categories"
	if val, ok := globalHEICatCache.Load(cacheKey); ok {
		item := val.(heiCatCacheItem)
		if time.Now().Before(item.expiresAt) {
			return item.categories, nil
		}
		globalHEICatCache.Delete(cacheKey)
	}

	var categories []string
	err := r.db.Model(&master.HEIMaster{}).
		Select("DISTINCT \"CategoryName\"").
		Where("\"Status\" = ?", "Active").
		Order("\"CategoryName\" ASC").
		Pluck("\"CategoryName\"", &categories).Error
	if err != nil {
		return nil, err
	}

	// Always ensure Habit, Equipment, Infrastructure are present
	defaultCats := []string{"Habit", "Equipment", "Infrastructure"}
	catMap := make(map[string]bool)
	for _, c := range defaultCats {
		catMap[c] = true
	}
	for _, c := range categories {
		if c != "" {
			catMap[c] = true
		}
	}

	var finalCats []string
	for _, c := range defaultCats {
		finalCats = append(finalCats, c)
		delete(catMap, c)
	}
	for c := range catMap {
		finalCats = append(finalCats, c)
	}

	globalHEICatCache.Store(cacheKey, heiCatCacheItem{
		categories: finalCats,
		expiresAt:  time.Now().Add(10 * time.Minute),
	})

	return finalCats, nil
}

func (r *heiRepository) Create(h *master.HEIMaster) error {
	globalHEICache = sync.Map{}
	globalHEICatCache = sync.Map{}
	return r.db.Create(h).Error
}

func (r *heiRepository) Update(h *master.HEIMaster) error {
	globalHEICache = sync.Map{}
	globalHEICatCache = sync.Map{}
	return r.db.Save(h).Error
}

func (r *heiRepository) Delete(id string) error {
	globalHEICache = sync.Map{}
	globalHEICatCache = sync.Map{}
	return r.db.Where("\"HEIID\" = ?", id).Delete(&master.HEIMaster{}).Error
}

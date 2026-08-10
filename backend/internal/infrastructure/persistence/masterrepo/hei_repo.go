package masterrepo

import (
	"strings"

	"github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

type heiRepository struct {
	db *gorm.DB
}

func NewHEIRepository(db *gorm.DB) master.HEIRepository {
	return &heiRepository{db: db}
}

func (r *heiRepository) FindAll(page, limit int, category, search string) ([]master.HEIMaster, int64, error) {
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

	out := make([]string, 0, len(catMap))
	for cat := range catMap {
		out = append(out, cat)
	}
	return out, nil
}

func (r *heiRepository) Create(h *master.HEIMaster) error {
	return r.db.Create(h).Error
}

func (r *heiRepository) Update(h *master.HEIMaster) error {
	return r.db.Save(h).Error
}

func (r *heiRepository) Delete(id string) error {
	return r.db.Where("\"HEIID\" = ?", id).Delete(&master.HEIMaster{}).Error
}

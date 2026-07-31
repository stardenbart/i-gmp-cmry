package masterrepo

import (
	"github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

type plantRepository struct {
	db *gorm.DB
}

func NewPlantRepository(db *gorm.DB) master.PlantRepository {
	return &plantRepository{db: db}
}

func (r *plantRepository) FindAll(page, limit int, search string) ([]master.Plant, int64, error) {
	var plants []master.Plant
	var total int64

	query := r.db.Model(&master.Plant{})
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("\"PlantName\" ILIKE ? OR \"PlantCode\" ILIKE ? OR \"Address\" ILIKE ?", s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("\"PlantCreatedAt\" DESC").Offset(offset).Limit(limit).Find(&plants).Error; err != nil {
		return nil, 0, err
	}

	return plants, total, nil
}

func (r *plantRepository) FindByID(id string) (*master.Plant, error) {
	var p master.Plant
	if err := r.db.Where("\"PlantID\" = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *plantRepository) FindByCode(code string) (*master.Plant, error) {
	var p master.Plant
	if err := r.db.Where("\"PlantCode\" = ?", code).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *plantRepository) Create(p *master.Plant) error {
	return r.db.Create(p).Error
}

func (r *plantRepository) Update(p *master.Plant) error {
	return r.db.Save(p).Error
}

func (r *plantRepository) Delete(id string) error {
	return r.db.Where("\"PlantID\" = ?", id).Delete(&master.Plant{}).Error
}

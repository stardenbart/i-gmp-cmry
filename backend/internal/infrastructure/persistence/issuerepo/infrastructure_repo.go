package issuerepo

import (
	"github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
)

type infrastructureRepository struct {
	db *gorm.DB
}

func NewInfrastructureRepository(db *gorm.DB) issue.InfrastructureRepository {
	return &infrastructureRepository{db: db}
}

func (r *infrastructureRepository) FindAll(page, limit int, kawasanID, search string) ([]issue.Infrastructure, int64, error) {
	var infrastructures []issue.Infrastructure
	var total int64

	query := r.db.Table("\"Infrastructure_Master\" i").
		Select("i.*, k.\"KawasanName\" as \"KawasanName\"").
		Joins("LEFT JOIN \"Kawasan_Master\" k ON k.\"KawasanID\" = i.\"KawasanID\"")

	if kawasanID != "" {
		query = query.Where("i.\"KawasanID\" = ?", kawasanID)
	}
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("i.\"InfrastructureName\" ILIKE ? OR i.\"InfrastructureCode\" ILIKE ? OR i.\"InfrastructureType\" ILIKE ?", s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("i.\"InfrastructureCreatedAt\" DESC").Offset(offset).Limit(limit).Find(&infrastructures).Error; err != nil {
		return nil, 0, err
	}

	return infrastructures, total, nil
}

func (r *infrastructureRepository) FindByID(id string) (*issue.Infrastructure, error) {
	var inf issue.Infrastructure
	err := r.db.Table("\"Infrastructure_Master\" i").
		Select("i.*, k.\"KawasanName\" as \"KawasanName\"").
		Joins("LEFT JOIN \"Kawasan_Master\" k ON k.\"KawasanID\" = i.\"KawasanID\"").
		Where("i.\"InfrastructureID\" = ?", id).First(&inf).Error
	if err != nil {
		return nil, err
	}
	return &inf, nil
}

func (r *infrastructureRepository) Create(inf *issue.Infrastructure) error {
	return r.db.Create(inf).Error
}

func (r *infrastructureRepository) Update(inf *issue.Infrastructure) error {
	return r.db.Save(inf).Error
}

func (r *infrastructureRepository) Delete(id string) error {
	return r.db.Where("\"InfrastructureID\" = ?", id).Delete(&issue.Infrastructure{}).Error
}

package issuerepo

import (
	"github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
)

type equipmentRepository struct {
	db *gorm.DB
}

func NewEquipmentRepository(db *gorm.DB) issue.EquipmentRepository {
	return &equipmentRepository{db: db}
}

func (r *equipmentRepository) FindAll(page, limit int, kawasanID, search string) ([]issue.Equipment, int64, error) {
	var equipments []issue.Equipment
	var total int64

	query := r.db.Table("\"Equipment_Master\" e").
		Select("e.*, k.\"KawasanName\" as \"KawasanName\"").
		Joins("LEFT JOIN \"Kawasan_Master\" k ON k.\"KawasanID\" = e.\"KawasanID\"")

	if kawasanID != "" {
		query = query.Where("e.\"KawasanID\" = ?", kawasanID)
	}
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("e.\"EquipmentName\" ILIKE ? OR e.\"EquipmentCode\" ILIKE ? OR e.\"EquipmentType\" ILIKE ?", s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("e.\"EquipmentCreatedAt\" DESC").Offset(offset).Limit(limit).Find(&equipments).Error; err != nil {
		return nil, 0, err
	}

	return equipments, total, nil
}

func (r *equipmentRepository) FindByID(id string) (*issue.Equipment, error) {
	var eq issue.Equipment
	err := r.db.Table("\"Equipment_Master\" e").
		Select("e.*, k.\"KawasanName\" as \"KawasanName\"").
		Joins("LEFT JOIN \"Kawasan_Master\" k ON k.\"KawasanID\" = e.\"KawasanID\"").
		Where("e.\"EquipmentID\" = ?", id).First(&eq).Error
	if err != nil {
		return nil, err
	}
	return &eq, nil
}

func (r *equipmentRepository) Create(eq *issue.Equipment) error {
	return r.db.Create(eq).Error
}

func (r *equipmentRepository) Update(eq *issue.Equipment) error {
	return r.db.Save(eq).Error
}

func (r *equipmentRepository) Delete(id string) error {
	return r.db.Where("\"EquipmentID\" = ?", id).Delete(&issue.Equipment{}).Error
}

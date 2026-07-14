package picrepo

import (
	"github.com/monitoring-system/backend/internal/domain/pic"
	"gorm.io/gorm"
)

type picMappingRepository struct{ db *gorm.DB }

func NewPICMappingRepository(db *gorm.DB) pic.PICMappingRepository {
	return &picMappingRepository{db: db}
}

func (r *picMappingRepository) FindAll(page, limit int, areaID, kawasanID string) ([]pic.PICMapping, int64, error) {
	var items []pic.PICMapping
	var total int64
	q := r.db.Model(&pic.PICMapping{})
	if areaID != "" { q = q.Where("AreaID = ?", areaID) }
	if kawasanID != "" { q = q.Where("KawasanID = ?", kawasanID) }
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *picMappingRepository) FindByID(id string) (*pic.PICMapping, error) {
	var item pic.PICMapping
	err := r.db.Where("PICMapID = ?", id).First(&item).Error
	return &item, err
}

func (r *picMappingRepository) FindByUserID(userID string) ([]pic.PICMapping, error) {
	var items []pic.PICMapping
	err := r.db.Where("UserID = ?", userID).Find(&items).Error
	return items, err
}

func (r *picMappingRepository) FindByAreaAndKawasan(areaID, kawasanID string) ([]pic.PICMapping, error) {
	var items []pic.PICMapping
	err := r.db.Where("AreaID = ? AND KawasanID = ?", areaID, kawasanID).Find(&items).Error
	return items, err
}

func (r *picMappingRepository) Create(p *pic.PICMapping) error   { return r.db.Create(p).Error }
func (r *picMappingRepository) Update(p *pic.PICMapping) error   { return r.db.Save(p).Error }
func (r *picMappingRepository) Delete(id string) error {
	return r.db.Where("PICMapID = ?", id).Delete(&pic.PICMapping{}).Error
}

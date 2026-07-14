package masterrepo

import (
	"github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

type settingRepository struct{ db *gorm.DB }

func NewSettingRepository(db *gorm.DB) master.SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) FindAll() ([]master.Setting, error) {
	var items []master.Setting
	return items, r.db.Order("\"SettingKey\"").Find(&items).Error
}

func (r *settingRepository) FindByKey(key string) (*master.Setting, error) {
	var item master.Setting
	return &item, r.db.Where("\"SettingKey\" = ?", key).First(&item).Error
}

func (r *settingRepository) Update(key, value, updatedBy string) error {
	return r.db.Model(&master.Setting{}).
		Where("\"SettingKey\" = ?", key).
		Updates(map[string]interface{}{
			"SettingValue": value,
			"UpdatedBy":    updatedBy,
		}).Error
}

package masterrepo

import (
	"github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

type settingRepository struct{ db *gorm.DB }

func NewSettingRepository(db *gorm.DB) master.SettingRepository {
	// Migration steps to transition to composite PK ("SettingKey", "PlantID")
	_ = db.Exec(`ALTER TABLE "System_Setting" ADD COLUMN IF NOT EXISTS "PlantID" varchar(50) NOT NULL DEFAULT ''`).Error
	_ = db.Exec(`UPDATE "System_Setting" SET "PlantID" = '' WHERE "PlantID" IS NULL`).Error
	_ = db.Exec(`ALTER TABLE "System_Setting" DROP CONSTRAINT IF EXISTS "System_Setting_pkey"`).Error
	_ = db.Exec(`ALTER TABLE "System_Setting" ADD CONSTRAINT "System_Setting_pkey" PRIMARY KEY ("SettingKey", "PlantID")`).Error

	_ = db.AutoMigrate(&master.Setting{})
	return &settingRepository{db: db}
}

func (r *settingRepository) FindAll(plantID string) ([]master.Setting, error) {
	var items []master.Setting
	q := r.db.Order("\"SettingKey\"")
	if plantID == "GLOBAL" || plantID == "NULL" {
		q = q.Where("\"PlantID\" = ''")
	} else if plantID != "" && plantID != "ALL" {
		q = q.Where("\"PlantID\" = ? OR \"PlantID\" = ''", plantID)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}

	// If plantID is a specific plant, filter duplicates so plant override takes precedence over global default
	if plantID != "" && plantID != "ALL" && plantID != "GLOBAL" && plantID != "NULL" {
		m := make(map[string]master.Setting)
		for _, s := range items {
			existing, exists := m[s.SettingKey]
			if !exists {
				m[s.SettingKey] = s
			} else if s.PlantID == plantID {
				m[s.SettingKey] = s
			} else if existing.PlantID == "" {
				// Keep current override
			}
		}
		res := make([]master.Setting, 0, len(m))
		for _, s := range m {
			res = append(res, s)
		}
		return res, nil
	}

	return items, nil
}

func (r *settingRepository) FindByKey(key, plantID string) (*master.Setting, error) {
	var item master.Setting
	if plantID != "" && plantID != "GLOBAL" && plantID != "NULL" {
		err := r.db.Where("\"SettingKey\" = ? AND \"PlantID\" = ?", key, plantID).First(&item).Error
		if err == nil {
			return &item, nil
		}
	}
	err := r.db.Where("\"SettingKey\" = ? AND \"PlantID\" = ''", key).First(&item).Error
	return &item, err
}

func (r *settingRepository) Update(key, value, updatedBy, plantID string) error {
	targetPlantID := plantID
	if targetPlantID == "GLOBAL" || targetPlantID == "NULL" {
		targetPlantID = ""
	}

	// Check if record exists for key + plantID
	var existing master.Setting
	err := r.db.Where("\"SettingKey\" = ? AND \"PlantID\" = ?", key, targetPlantID).First(&existing).Error
	if err == nil {
		// Update existing
		return r.db.Model(&master.Setting{}).
			Where("\"SettingKey\" = ? AND \"PlantID\" = ?", key, targetPlantID).
			Updates(map[string]interface{}{
				"SettingValue": value,
				"UpdatedBy":    updatedBy,
			}).Error
	}

	// Get description from global default if available
	var globalDef master.Setting
	desc := ""
	if r.db.Where("\"SettingKey\" = ? AND \"PlantID\" = ''", key).First(&globalDef).Error == nil {
		desc = globalDef.Description
	}

	// Create new override
	newSetting := master.Setting{
		SettingKey:   key,
		PlantID:      targetPlantID,
		SettingValue: value,
		Description:  desc,
	}
	if updatedBy != "" {
		newSetting.UpdatedBy = &updatedBy
	}

	return r.db.Create(&newSetting).Error
}

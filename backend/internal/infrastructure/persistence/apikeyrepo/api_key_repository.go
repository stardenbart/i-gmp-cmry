package apikeyrepo

import (
	"errors"
	"time"

	"github.com/monitoring-system/backend/internal/domain/apikey"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) apikey.Repository {
	// AutoMigrate API_Key table
	if err := db.AutoMigrate(&apikey.APIKey{}); err != nil {
		panic("failed to automigrate API_Key: " + err.Error())
	}
	return &repository{db: db}
}

func (r *repository) Create(key *apikey.APIKey) error {
	return r.db.Create(key).Error
}

func (r *repository) FindAll(createdBy, plantID string) ([]apikey.APIKey, error) {
	var keys []apikey.APIKey
	query := r.db
	if createdBy != "" {
		query = query.Where("\"CreatedByID\" = ?", createdBy)
	}
	if plantID != "" && plantID != "ALL" {
		if plantID == "GLOBAL" || plantID == "NULL" {
			query = query.Where("\"PlantID\" IS NULL OR \"PlantID\" = ''")
		} else {
			query = query.Where("\"PlantID\" = ?", plantID)
		}
	}
	err := query.Order(`"CreatedAt" DESC`).Find(&keys).Error
	return keys, err
}

func (r *repository) FindByID(keyID string) (*apikey.APIKey, error) {
	var k apikey.APIKey
	// Struct-based Where — GORM uses the column:KeyID tag automatically
	err := r.db.Where(&apikey.APIKey{KeyID: keyID}).First(&k).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

func (r *repository) FindByHash(keyHash string) (*apikey.APIKey, error) {
	var k apikey.APIKey
	// Struct-based Where — GORM uses the column:KeyHash tag automatically
	err := r.db.Where(&apikey.APIKey{KeyHash: keyHash}).First(&k).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

func (r *repository) MarkAsUsed(keyID string) error {
	now := time.Now()
	return r.db.Model(&apikey.APIKey{}).
		Where(&apikey.APIKey{KeyID: keyID}).
		Updates(map[string]interface{}{
			"IsActive": false,
			"UsedAt":   now,
		}).Error
}

func (r *repository) Revoke(keyID string) error {
	return r.db.Model(&apikey.APIKey{}).
		Where(&apikey.APIKey{KeyID: keyID}).
		Update("IsActive", false).Error
}

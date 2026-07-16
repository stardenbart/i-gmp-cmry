package uploadrepo

import (
	"gorm.io/gorm"

	"github.com/monitoring-system/backend/internal/domain/upload"
)

// UploadRepository handles database operations for uploads
type UploadRepository struct {
	db *gorm.DB
}

// NewUploadRepository creates a new upload repository instance
func NewUploadRepository(db *gorm.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

// Create inserts a new upload record
func (r *UploadRepository) Create(upload *upload.Upload) error {
	return r.db.Create(upload).Error
}

// GetByID retrieves an upload by its ID
func (r *UploadRepository) GetByID(id string) (*upload.Upload, error) {
	var u upload.Upload
	if err := r.db.Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByInspectionID retrieves all uploads for a given inspection
func (r *UploadRepository) GetByInspectionID(inspectionID string) ([]upload.Upload, error) {
	var uploads []upload.Upload
	if err := r.db.Where("inspection_id = ?", inspectionID).Find(&uploads).Error; err != nil {
		return nil, err
	}
	return uploads, nil
}

// UpdateStatus updates the status and processed URL of an upload
func (r *UploadRepository) UpdateStatus(id string, status string, processedURL *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if processedURL != nil {
		updates["processed_url"] = *processedURL
	}
	return r.db.Model(&upload.Upload{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateProcessedURL updates only the processed URL field
func (r *UploadRepository) UpdateProcessedURL(id string, processedURL string) error {
	return r.db.Model(&upload.Upload{}).Where("id = ?", id).Update("processed_url", processedURL).Error
}

// Delete removes an upload record
func (r *UploadRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&upload.Upload{}).Error
}

// List retrieves all uploads with pagination
func (r *UploadRepository) List(page, limit int) ([]upload.Upload, int64, error) {
	var uploads []upload.Upload
	var total int64

	if err := r.db.Model(&upload.Upload{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&uploads).Error; err != nil {
		return nil, 0, err
	}

	return uploads, total, nil
}

// GetProcessingUploads retrieves uploads with 'processing' status
func (r *UploadRepository) GetProcessingUploads() ([]upload.Upload, error) {
	var uploads []upload.Upload
	if err := r.db.Where("status = ?", upload.StatusProcessing).Find(&uploads).Error; err != nil {
		return nil, err
	}
	return uploads, nil
}

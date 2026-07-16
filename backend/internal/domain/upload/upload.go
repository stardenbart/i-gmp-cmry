package upload

import "errors"

// Content type constants
const (
	ContentTypeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
	ContentTypeGIF  = "image/gif"
	ContentTypeWebP = "image/webp"
)

// Allowed extensions
var AllowedDOCXExtensions = []string{".docx"}
var AllowedImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif"}

// Max file sizes
const (
	MaxDOCXSize  = 50 * 1024 * 1024 // 50MB
	MaxImageSize = 10 * 1024 * 1024 // 10MB
)

// Kafka topics
const (
	TopicImageProcessing = "audit.image-processing"
)

// File type constants
const (
	FileTypeDOCX  = "docx"
	FileTypeImage = "image"
)

// Status constants
const (
	StatusProcessing = "processing"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// Common errors
var (
	ErrInvalidExtension  = errors.New("file extension not allowed")
	ErrInvalidMIMEType  = errors.New("invalid file type")
	ErrFileTooLarge      = errors.New("file size exceeds maximum")
	ErrCorruptedDOCX     = errors.New("corrupted or invalid docx file")
	ErrInvalidMagicBytes = errors.New("invalid magic bytes")
)

// UploadRequest - Request body untuk upload
type UploadRequest struct {
	FileType     string `json:"file_type" validate:"required,oneof=docx image"`
	InspectionID string `json:"inspection_id" validate:"required,uuid"`
}

// UploadResponse - Response setelah upload berhasil
type UploadResponse struct {
	ID           string `json:"id"`
	FileName     string `json:"file_name"`
	FileURL      string `json:"file_url"`
	FileSize     int64  `json:"file_size"`
	ContentType  string `json:"content_type"`
	Status       string `json:"status"`
	ProcessedURL string `json:"processed_url,omitempty"`
}

// ImageProcessingMessage - Kafka message untuk image processing
type ImageProcessingMessage struct {
	FileID       string `json:"file_id"`
	SourceURL    string `json:"source_url"`
	ObjectName   string `json:"object_name"`
}

// Upload - Database model
type Upload struct {
	ID               string  `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	InspectionID     string  `gorm:"column:inspection_id;type:uuid;not null"`
	OriginalFilename string  `gorm:"column:original_filename;type:varchar(255);not null"`
	StoredFilename   string  `gorm:"column:stored_filename;type:varchar(255);not null"`
	FilePath         string  `gorm:"column:file_path;type:text;not null"`
	FileSize         int64   `gorm:"column:file_size;not null"`
	ContentType      string  `gorm:"column:content_type;type:varchar(100);not null"`
	FileType         string  `gorm:"column:file_type;type:varchar(20);not null"`
	Status           string  `gorm:"column:status;type:varchar(20);not null;default:'completed'"`
	ProcessedURL     *string `gorm:"column:processed_url;type:text"`
	CreatedAt        int64   `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        int64   `gorm:"column:updated_at;autoCreateTime:milli"`
}

// TableName specifies the table name for the Upload model
func (Upload) TableName() string {
	return "uploads"
}

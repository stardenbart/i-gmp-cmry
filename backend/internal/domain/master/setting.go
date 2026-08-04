package master

import "time"

// Setting represents a dynamic system configuration key-value pair.
type Setting struct {
	SettingKey   string    `gorm:"column:SettingKey;primaryKey" json:"setting_key"`
	PlantID      *string   `gorm:"column:PlantID" json:"plant_id,omitempty"`
	SettingValue string    `gorm:"column:SettingValue;not null" json:"setting_value"`
	IsEncrypted  bool      `gorm:"column:IsEncrypted;not null;default:false" json:"is_encrypted"`
	Description  string    `gorm:"column:Description;size:255" json:"description"`
	UpdatedAt    time.Time `gorm:"column:UpdatedAt;autoUpdateTime" json:"updated_at"`
	UpdatedBy    *string   `gorm:"column:UpdatedBy" json:"updated_by,omitempty"`
}

func (Setting) TableName() string { return "System_Setting" }

// ── Well-known Setting Keys ────────────────────────────────────────────────

const (
	SettingKeyMinioAllowedIPs                  = "MINIO_ALLOWED_IPS"
	SettingKeyMaxUploadSizeMB                  = "MAX_UPLOAD_SIZE_MB"
	SettingKeyMaxLoginAttempts                 = "MAX_LOGIN_ATTEMPTS"
	SettingKeySessionIdleTimeout               = "SESSION_IDLE_TIMEOUT_MINUTES"
	SettingKeyEmailTemplateForgotPass          = "EMAIL_TEMPLATE_FORGOT_PASSWORD"
	SettingKeyEmailTemplateIssue               = "EMAIL_TEMPLATE_ISSUE_ASSIGNMENT"
	SettingKeyEmailTemplateInspectionConfirmed = "EMAIL_TEMPLATE_INSPECTION_CONFIRMED"
	SettingKeyIssueDeadlineDays                = "ISSUE_DEADLINE_DAYS"
	SettingKeyIssueAutoApproveDays             = "ISSUE_AUTO_APPROVE_DAYS"
)

// ── DTOs ──────────────────────────────────────────────────────────────────

type UpdateSettingRequest struct {
	SettingValue string  `json:"setting_value" validate:"required"`
	PlantID      *string `json:"plant_id,omitempty"`
}

// SettingResponse is a sanitized view that never exposes raw encrypted values.
type SettingResponse struct {
	SettingKey   string    `json:"setting_key"`
	PlantID      *string   `json:"plant_id,omitempty"`
	SettingValue string    `json:"setting_value"` // will be masked if IsEncrypted
	IsEncrypted  bool      `json:"is_encrypted"`
	Description  string    `json:"description"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    *string   `json:"updated_by,omitempty"`
}

// ── Repository Interface ──────────────────────────────────────────────────

type SettingRepository interface {
	FindAll(plantID string) ([]Setting, error)
	FindByKey(key, plantID string) (*Setting, error)
	Update(key, value, updatedBy, plantID string) error
}

// ── UseCase Interface ─────────────────────────────────────────────────────

type SettingUseCase interface {
	GetAll(plantID string) ([]SettingResponse, error)
	GetByKey(key, plantID string) (*SettingResponse, error)
	Update(key string, req *UpdateSettingRequest, updatedBy, plantID string) (*SettingResponse, error)
}

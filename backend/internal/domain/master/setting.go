package master

import "time"

// Setting represents a dynamic system configuration key-value pair.
type Setting struct {
	SettingKey   string     `gorm:"column:SettingKey;primaryKey" json:"setting_key"`
	SettingValue string     `gorm:"column:SettingValue;not null" json:"setting_value"`
	IsEncrypted  bool       `gorm:"column:IsEncrypted;not null;default:false" json:"is_encrypted"`
	Description  string     `gorm:"column:Description;size:255" json:"description"`
	UpdatedAt    time.Time  `gorm:"column:UpdatedAt;autoUpdateTime" json:"updated_at"`
	UpdatedBy    *string    `gorm:"column:UpdatedBy" json:"updated_by,omitempty"`
}

func (Setting) TableName() string { return "System_Setting" }

// ── Well-known Setting Keys ────────────────────────────────────────────────

const (
	SettingKeyMinioAllowedIPs         = "MINIO_ALLOWED_IPS"
	SettingKeyMaxUploadSizeMB         = "MAX_UPLOAD_SIZE_MB"
	SettingKeyMaxLoginAttempts        = "MAX_LOGIN_ATTEMPTS"
	SettingKeySessionIdleTimeout      = "SESSION_IDLE_TIMEOUT_MINUTES"
	SettingKeyEmailTemplateForgotPass = "EMAIL_TEMPLATE_FORGOT_PASSWORD"
	SettingKeyEmailTemplateIssue      = "EMAIL_TEMPLATE_ISSUE_ASSIGNMENT"
)

// ── DTOs ──────────────────────────────────────────────────────────────────

type UpdateSettingRequest struct {
	SettingValue string `json:"setting_value" validate:"required"`
}

// SettingResponse is a sanitized view that never exposes raw encrypted values.
type SettingResponse struct {
	SettingKey  string    `json:"setting_key"`
	SettingValue string   `json:"setting_value"` // will be masked if IsEncrypted
	IsEncrypted bool      `json:"is_encrypted"`
	Description  string   `json:"description"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    *string   `json:"updated_by,omitempty"`
}

// ── Repository Interface ──────────────────────────────────────────────────

type SettingRepository interface {
	FindAll() ([]Setting, error)
	FindByKey(key string) (*Setting, error)
	Update(key, value string, updatedBy string) error
}

// ── UseCase Interface ─────────────────────────────────────────────────────

type SettingUseCase interface {
	GetAll() ([]SettingResponse, error)
	GetByKey(key string) (*SettingResponse, error)
	Update(key string, req *UpdateSettingRequest, updatedBy string) (*SettingResponse, error)
}

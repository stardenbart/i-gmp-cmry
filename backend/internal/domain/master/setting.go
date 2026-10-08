package master

import "time"

// Setting represents a dynamic system configuration key-value pair.
type Setting struct {
	SettingKey   string    `gorm:"column:SettingKey;primaryKey" json:"setting_key"`
	PlantID      string    `gorm:"column:PlantID;primaryKey;default:''" json:"plant_id"`
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
	SettingKeyEmailTemplateKawasanConfirmed    = "EMAIL_TEMPLATE_KAWASAN_CONFIRMED"
	// SettingKeyEmailCCKawasanReport lists the addresses (separated by ";",
	// "," or new lines) copied on every Kawasan inspection report email.
	SettingKeyEmailCCKawasanReport = "EMAIL_CC_KAWASAN_REPORT"
	SettingKeyEmailTemplateDeadlineReminder    = "EMAIL_TEMPLATE_DEADLINE_REMINDER"
	SettingKeyIssueDeadlineDays                = "ISSUE_DEADLINE_DAYS"
	SettingKeyIssueAutoApproveDays             = "ISSUE_AUTO_APPROVE_DAYS"
	SettingKeyIssueWOWRAutoApproveDays         = "ISSUE_WOWR_AUTO_APPROVE_DAYS"
	// SettingKeyIssueDeadlineReminderDaysBefore controls how many days before
	// an Issue's DueDate DeadlineReminderWorker sends a reminder. "0" (or
	// unset, via the 0-default in readDaysSetting-style resolution) disables
	// the reminder entirely for that plant.
	SettingKeyIssueDeadlineReminderDaysBefore = "ISSUE_DEADLINE_REMINDER_DAYS_BEFORE"
	// SettingKeyNotifyOnKawasanComplete / SettingKeyNotifyOnAreaComplete gate
	// the two "fully inspected" email notifications independently ("true"/
	// "false"; treated as enabled when unset, matching prior behavior before
	// these toggles existed).
	SettingKeyNotifyOnKawasanComplete = "NOTIFY_ON_KAWASAN_COMPLETE"
	SettingKeyNotifyOnAreaComplete    = "NOTIFY_ON_AREA_COMPLETE"
	// SettingKeyInspectionPeriodCutoffDay controls which day of the month
	// starts a new "inspection period" (the recurring monthly cycle a
	// DetailKawasan must be re-inspected within) — e.g. "13" means the
	// period 13 Jan-12 Feb counts as "January". Per-plant, with the usual
	// plant→global fallback; "1" (the default) is an exact calendar month,
	// so any plant that never sets this keeps its current behavior exactly.
	SettingKeyInspectionPeriodCutoffDay = "INSPECTION_PERIOD_CUTOFF_DAY"
	SettingKeySMTPEnabled               = "SMTP_ENABLED"
	SettingKeySMTPHost                  = "SMTP_HOST"
	SettingKeySMTPPort                  = "SMTP_PORT"
	SettingKeySMTPUser                  = "SMTP_USER"
	SettingKeySMTPPassword              = "SMTP_PASSWORD"
	SettingKeySMTPSenderEmail           = "SMTP_SENDER_EMAIL"
)

func IsSMTPSettingKey(key string) bool {
	switch key {
	case SettingKeySMTPEnabled, SettingKeySMTPHost, SettingKeySMTPPort,
		SettingKeySMTPUser, SettingKeySMTPPassword, SettingKeySMTPSenderEmail:
		return true
	default:
		return false
	}
}

// ── DTOs ──────────────────────────────────────────────────────────────────

type UpdateSettingRequest struct {
	SettingValue string `json:"setting_value" validate:"required"`
	PlantID      string `json:"plant_id,omitempty"`
}

// SettingResponse is a sanitized view that never exposes raw encrypted values.
type SettingResponse struct {
	SettingKey   string    `json:"setting_key"`
	PlantID      string    `json:"plant_id,omitempty"`
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

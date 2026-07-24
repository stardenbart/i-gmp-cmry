package logging

import "time"

// LoginStatus defines the result of a login attempt.
type LoginStatus string

const (
	LoginStatusSuccess LoginStatus = "Success"
	LoginStatusFailed  LoginStatus = "Failed"
)

// LoginLog represents the Login_Log table.
// Records every login/logout event.
type LoginLog struct {
	LoginLogID  string      `gorm:"column:LoginLogID;primaryKey" json:"login_log_id"`
	UserID      *string     `gorm:"column:UserID" json:"user_id,omitempty"`
	LoginAt     time.Time   `gorm:"column:LoginAt;not null" json:"login_at"`
	LogoutAt    *time.Time  `gorm:"column:LogoutAt" json:"logout_at,omitempty"`
	IPAddress   string      `gorm:"column:IPAddress;size:50" json:"ip_address"`
	DeviceInfo  string      `gorm:"column:DeviceInfo;size:255" json:"device_info"`
	LoginStatus LoginStatus `gorm:"column:LoginStatus;size:20;not null" json:"login_status"`
}

func (LoginLog) TableName() string { return "Login_Log" }

// ─── Repository Interface ──────────────────────────────────────────────────

type LoginLogRepository interface {
	FindAll(page, limit int, userID string) ([]LoginLog, int64, error)
	FindByID(id string) (*LoginLog, error)
	Create(l *LoginLog) error
	// UpdateLogoutAt records the logout timestamp for a session.
	UpdateLogoutAt(id string, logoutAt time.Time) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type LoginLogUseCase interface {
	GetAll(page, limit int, userID string) ([]LoginLog, int64, error)
	GetByID(id string) (*LoginLog, error)
}

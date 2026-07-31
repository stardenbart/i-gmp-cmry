package inspection

import (
	"time"
)

// KawasanAspek represents the mapping between Kawasan and Aspek
type KawasanAspek struct {
	KawasanAspekID string    `gorm:"column:KawasanAspekID;primaryKey" json:"kawasan_aspek_id"`
	KawasanID       string    `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	AspekID         string    `gorm:"column:AspekID;not null" json:"aspek_id"`
	IsRequired      bool      `gorm:"column:IsRequired;default:true" json:"is_required"`
	CreatedAt       time.Time `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
}

func (KawasanAspek) TableName() string {
	return "kawasan_aspek"
}

// InspeksiSession represents an inspection session for a Kawasan per period
type InspeksiSession struct {
	SessionID      string     `gorm:"column:SessionID;primaryKey" json:"session_id"`
	KawasanID      string     `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	Periode        time.Time  `gorm:"column:Periode;type:date;not null" json:"periode"`
	Status         string     `gorm:"column:Status;default:'InProgress'" json:"status"` // InProgress | Completed | Synced | Failed
	TotalAspek     int        `gorm:"column:TotalAspek;default:0" json:"total_aspek"`
	CompletedAspek int        `gorm:"column:CompletedAspek;default:0" json:"completed_aspek"`
	SyncedAt       *time.Time `gorm:"column:SyncedAt" json:"synced_at,omitempty"`
	CreatedAt      time.Time  `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:UpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (InspeksiSession) TableName() string {
	return "inspeksi_session"
}

// InspeksiAspekResult represents the completed inspection result per aspek
type InspeksiAspekResult struct {
	ResultID     string    `gorm:"column:ResultID;primaryKey" json:"result_id"`
	SessionID    string    `gorm:"column:SessionID;not null" json:"session_id"`
	KawasanID    string    `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	AspekID      string    `gorm:"column:AspekID;not null" json:"aspek_id"`
	UserID       string    `gorm:"column:UserID;not null" json:"user_id"`
	DataInspeksi string    `gorm:"column:DataInspeksi;type:jsonb;not null" json:"data_inspeksi"` // JSON payload string
	Skor         float64   `gorm:"column:Skor" json:"skor"`
	Catatan      string    `gorm:"column:Catatan" json:"catatan"`
	CompletedAt  time.Time `gorm:"column:CompletedAt;not null" json:"completed_at"`
	CreatedAt    time.Time `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
}

func (InspeksiAspekResult) TableName() string {
	return "inspeksi_aspek_result"
}

// InspeksiAuditLog records lock/unlock/completion activities
type InspeksiAuditLog struct {
	LogID     string    `gorm:"column:LogID;primaryKey" json:"log_id"`
	SessionID string    `gorm:"column:SessionID;not null" json:"session_id"`
	AspekID   string    `gorm:"column:AspekID;not null" json:"aspek_id"`
	UserID    string    `gorm:"column:UserID;not null" json:"user_id"`
	Action    string    `gorm:"column:Action;not null" json:"action"`
	Meta      string    `gorm:"column:Meta;type:jsonb" json:"meta"`
	CreatedAt time.Time `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
}

func (InspeksiAuditLog) TableName() string {
	return "inspeksi_audit_log"
}

// DTOs & Real-time structs
type LockResult struct {
	LockToken string    `json:"lock_token"`
	ExpiresAt time.Time `json:"expires_at"`
	LockedBy  string    `json:"locked_by,omitempty"`
}

type AspekLockStatus struct {
	AspekID       string    `json:"aspek_id"`
	AspekName     string    `json:"aspek_name,omitempty"`
	Status        string    `json:"status"` // FREE | LOCKED | COMPLETED
	LockedBy      string    `json:"locked_by,omitempty"`
	LockedByName  string    `json:"locked_by_name,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CompletedBy   string    `json:"completed_by,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

type SaveAspekRequest struct {
	KawasanID string                 `json:"kawasan_id" validate:"required"`
	AspekID   string                 `json:"aspek_id" validate:"required"`
	SessionID string                 `json:"session_id" validate:"required"`
	UserID    string                 `json:"user_id"`
	LockToken string                 `json:"lock_token" validate:"required"`
	Data      map[string]interface{} `json:"data" validate:"required"`
	Skor      float64                `json:"skor"`
	IsFinal   bool                   `json:"is_final"`
}

type YieldRequestPayload struct {
	KawasanID string `json:"kawasan_id"`
	AspekID   string `json:"aspek_id"`
	Reason    string `json:"reason,omitempty"`
}

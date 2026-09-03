package kpishare

import "time"

type PublicShare struct {
	ShareID           string     `gorm:"column:ShareID;primaryKey;size:50" json:"share_id"`
	OwnerUserID       string     `gorm:"column:OwnerUserID;size:50;not null" json:"owner_user_id"`
	ShareName         string     `gorm:"column:ShareName;size:100;not null" json:"share_name"`
	PublicTitle       string     `gorm:"column:PublicTitle;size:150;not null" json:"public_title"`
	TokenHash         string     `gorm:"column:TokenHash;size:64;not null;uniqueIndex" json:"-"`
	TokenPrefix       string     `gorm:"column:TokenPrefix;size:16;not null" json:"token_prefix"`
	PlantID           string     `gorm:"column:PlantID;size:50;not null" json:"plant_id"`
	LayoutJSON        string     `gorm:"column:LayoutJSON;type:jsonb;not null" json:"-"`
	FilterJSON        string     `gorm:"column:FilterJSON;type:jsonb;not null" json:"-"`
	AllowPeriodChange bool       `gorm:"column:AllowPeriodChange;not null" json:"allow_period_change"`
	SnapshotVersion   int16      `gorm:"column:SnapshotVersion;not null" json:"snapshot_version"`
	ExpiresAt         *time.Time `gorm:"column:ExpiresAt" json:"expires_at,omitempty"`
	RevokedAt         *time.Time `gorm:"column:RevokedAt" json:"revoked_at,omitempty"`
	LastAccessedAt    *time.Time `gorm:"column:LastAccessedAt" json:"last_accessed_at,omitempty"`
	AccessCount       int64      `gorm:"column:AccessCount;not null" json:"access_count"`
	CreatedAt         time.Time  `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:UpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (PublicShare) TableName() string { return "KPI_Public_Share" }

type FilterSnapshot struct {
	Period      string `json:"period"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
	Granularity string `json:"granularity,omitempty"`
}

type Actor struct {
	UserID  string
	RoleID  string
	PlantID string
}

type CreateRequest struct {
	ShareName         string         `json:"share_name"`
	PublicTitle       string         `json:"public_title"`
	PlantID           string         `json:"plant_id"`
	Filter            FilterSnapshot `json:"filter"`
	AllowPeriodChange bool           `json:"allow_period_change"`
	ExpiresAt         *time.Time     `json:"expires_at"`
}

type CreateResponse struct {
	ShareResponse
	RawToken string `json:"raw_token"`
}

type ShareResponse struct {
	ShareID           string     `json:"share_id"`
	ShareName         string     `json:"share_name"`
	PublicTitle       string     `json:"public_title"`
	TokenPrefix       string     `json:"token_prefix"`
	PlantID           string     `json:"plant_id"`
	PlantName         string     `json:"plant_name"`
	AllowPeriodChange bool       `json:"allow_period_change"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	LastAccessedAt    *time.Time `json:"last_accessed_at,omitempty"`
	AccessCount       int64      `json:"access_count"`
	CreatedAt         time.Time  `json:"created_at"`
}

type BootstrapResponse struct {
	ShareID           string         `json:"share_id"`
	PublicTitle       string         `json:"public_title"`
	PlantID           string         `json:"plant_id"`
	PlantName         string         `json:"plant_name"`
	Layout            interface{}    `json:"layout"`
	Filter            FilterSnapshot `json:"filter"`
	AllowPeriodChange bool           `json:"allow_period_change"`
	ExpiresAt         *time.Time     `json:"expires_at,omitempty"`
	UpdatedAt         time.Time      `json:"updated_at"`
	Catalog           interface{}    `json:"catalog,omitempty"`
}

type Repository interface {
	Create(share *PublicShare) error
	ListByOwner(ownerUserID string) ([]PublicShare, error)
	FindByID(shareID string) (*PublicShare, error)
	FindByTokenHash(tokenHash string) (*PublicShare, error)
	Revoke(shareID, ownerUserID string, now time.Time) (bool, error)
	Rotate(shareID, ownerUserID, tokenHash, tokenPrefix string, now time.Time) (*PublicShare, error)
	TouchAccess(shareID string, now time.Time) error
	PlantName(plantID string) (string, error)
}

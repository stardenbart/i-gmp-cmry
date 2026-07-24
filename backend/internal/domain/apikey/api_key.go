package apikey

import (
	"time"
)

// APIKey represents a key created by Admin to access public endpoints (e.g. Power BI)
type APIKey struct {
	KeyID       string     `gorm:"column:KeyID;primaryKey;size:50" json:"key_id"`
	Name        string     `gorm:"column:Name;size:100;not null" json:"name"`
	KeyHash     string     `gorm:"column:KeyHash;size:255;uniqueIndex;not null" json:"-"`
	Prefix      string     `gorm:"column:Prefix;size:20;not null" json:"prefix"`
	CreatedByID string     `gorm:"column:CreatedByID;size:50;not null" json:"created_by_id"`
	IsActive    bool       `gorm:"column:IsActive;default:true" json:"is_active"`
	IsSingleUse bool       `gorm:"column:IsSingleUse;default:true" json:"is_single_use"`
	UsedAt      *time.Time `gorm:"column:UsedAt" json:"used_at,omitempty"`
	ExpiresAt   *time.Time `gorm:"column:ExpiresAt" json:"expires_at,omitempty"`
	CreatedAt   time.Time  `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
}

func (APIKey) TableName() string {
	return "API_Key"
}

// DTOs
type CreateAPIKeyRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	IsSingleUse *bool  `json:"is_single_use"`
}

type CreateAPIKeyResponse struct {
	KeyID       string    `json:"key_id"`
	Name        string    `json:"name"`
	RawToken    string    `json:"raw_token"` // Returned ONLY once upon creation
	Prefix      string    `json:"prefix"`
	IsSingleUse bool      `json:"is_single_use"`
	CreatedAt   time.Time `json:"created_at"`
}

type APIKeyResponse struct {
	KeyID       string     `json:"key_id"`
	Name        string     `json:"name"`
	Prefix      string     `json:"prefix"`
	CreatedByID string     `json:"created_by_id"`
	IsActive    bool       `json:"is_active"`
	IsSingleUse bool       `json:"is_single_use"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Repository Interface
type Repository interface {
	Create(key *APIKey) error
	FindAll(createdBy string) ([]APIKey, error)
	FindByID(keyID string) (*APIKey, error)
	FindByHash(keyHash string) (*APIKey, error)
	MarkAsUsed(keyID string) error
	Revoke(keyID string) error
}

// UseCase Interface
type UseCase interface {
	CreateKey(req *CreateAPIKeyRequest, createdByID string) (*CreateAPIKeyResponse, error)
	ListKeys(createdBy string) ([]APIKeyResponse, error)
	RevokeKey(keyID string, requestedBy string) error
	ValidateAndConsumeKey(rawToken string) (*APIKey, error)
}

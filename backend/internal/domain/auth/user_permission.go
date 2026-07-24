package auth

import "time"

// UserPermission represents the User_Permission table for user-level overrides.
type UserPermission struct {
	UserPermissionID        string    `gorm:"column:UserPermissionID;primaryKey" json:"user_permission_id"`
	UserID                  string    `gorm:"column:UserID;not null" json:"user_id"`
	PermissionID            string    `gorm:"column:PermissionID;not null" json:"permission_id"`
	IsAllowed               bool      `gorm:"column:IsAllowed;not null" json:"is_allowed"`
	UserPermissionCreatedAt time.Time `gorm:"column:UserPermissionCreatedAt;autoCreateTime" json:"created_at"`
	UserPermissionUpdatedAt time.Time `gorm:"column:UserPermissionUpdatedAt;autoUpdateTime" json:"updated_at"`
	UserPermissionUpdatedBy string    `gorm:"column:UserPermissionUpdatedBy" json:"updated_by"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID;references:UserID" json:"user,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID;references:PermissionID" json:"permission,omitempty"`
}

func (UserPermission) TableName() string { return "User_Permission" }

// BulkSetUserPermissionRequest allows admin to set overrides for a user.
type BulkSetUserPermissionRequest struct {
	UserID      string             `json:"user_id" validate:"required"`
	Permissions []PermissionToggle `json:"permissions" validate:"required,dive"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type UserPermissionRepository interface {
	FindByUserID(userID string) ([]UserPermission, error)
	// CheckOverride returns (*bool, error). 
	// If the user has an explicit override, it returns a pointer to the bool.
	// If no override exists, it returns (nil, nil).
	CheckOverride(userID, moduleID, permissionCode string) (*bool, error)
	BulkUpsert(ups []UserPermission) error
	Delete(id string) error
	DeleteByUserID(userID string) error // to clear overrides
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type UserPermissionUseCase interface {
	GetByUserID(userID string) ([]UserPermission, error)
	SetPermissionOverrides(updatedByUserID string, req *BulkSetUserPermissionRequest) error
	// CheckOverride checks if a user has an explicit permission override.
	CheckOverride(userID, moduleID, permissionCode string) (*bool, error)
	ClearOverrides(userID string) error
}

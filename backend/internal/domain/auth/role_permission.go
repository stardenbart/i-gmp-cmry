package auth

import "time"

// RolePermission represents the Role_Permission table.
// This is the core of the dynamic RBAC system — admin can configure
// which permissions each role has without code changes.
type RolePermission struct {
	RolePermissionID        string    `gorm:"column:RolePermissionID;primaryKey" json:"role_permission_id"`
	RoleID                  string    `gorm:"column:RoleID;not null" json:"role_id"`
	PermissionID            string    `gorm:"column:PermissionID;not null" json:"permission_id"`
	IsAllowed               bool      `gorm:"column:IsAllowed;not null;default:true" json:"is_allowed"`
	RolePermissionCreatedAt time.Time `gorm:"column:RolePermissionCreatedAt;autoCreateTime" json:"created_at"`
	RolePermissionUpdatedAt time.Time `gorm:"column:RolePermissionUpdatedAt;autoUpdateTime" json:"updated_at"`
	RolePermissionUpdatedBy string    `gorm:"column:RolePermissionUpdatedBy" json:"updated_by"`

	// Relations
	Role       *Role       `gorm:"foreignKey:RoleID;references:RoleID" json:"role,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID;references:PermissionID" json:"permission,omitempty"`
}

func (RolePermission) TableName() string { return "Role_Permission" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

// BulkSetRequest allows admin to set all permissions for a role at once.
type BulkSetPermissionRequest struct {
	RoleID      string                      `json:"role_id" validate:"required"`
	Permissions []PermissionToggle          `json:"permissions" validate:"required,dive"`
}

type PermissionToggle struct {
	PermissionID string `json:"permission_id" validate:"required"`
	IsAllowed    bool   `json:"is_allowed"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type RolePermissionRepository interface {
	FindByRoleID(roleID string) ([]RolePermission, error)
	FindByRoleAndPermission(roleID, permissionID string) (*RolePermission, error)
	// HasPermission checks if roleID has a specific permissionCode on a moduleID.
	HasPermission(roleID, moduleID, permissionCode string) (bool, error)
	Upsert(rp *RolePermission) error
	BulkUpsert(rps []RolePermission) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type RolePermissionUseCase interface {
	GetByRoleID(roleID string) ([]RolePermission, error)
	SetPermission(updatedByUserID string, req *BulkSetPermissionRequest) error
	CheckPermission(roleID, moduleID, permissionCode string) (bool, error)
}

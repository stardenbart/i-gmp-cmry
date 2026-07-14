package auth

import "time"

// PermissionCode defines standard permission action codes.
type PermissionCode string

const (
	PermCodeCreate  PermissionCode = "CREATE"
	PermCodeRead    PermissionCode = "READ"
	PermCodeUpdate  PermissionCode = "UPDATE"
	PermCodeDelete  PermissionCode = "DELETE"
	PermCodeApprove PermissionCode = "APPROVE"
	PermCodeExport  PermissionCode = "EXPORT"
)

// Permission represents the Permission_Master table.
type Permission struct {
	PermissionID        string    `gorm:"column:PermissionID;primaryKey" json:"permission_id"`
	ModuleID            string    `gorm:"column:ModuleID;not null" json:"module_id"`
	PermissionCode      string    `gorm:"column:PermissionCode;not null" json:"permission_code"`
	PermissionName      string    `gorm:"column:PermissionName;not null" json:"permission_name"`
	PermissionCreatedAt time.Time `gorm:"column:PermissionCreatedAt;autoCreateTime" json:"created_at"`
	PermissionUpdatedAt time.Time `gorm:"column:PermissionUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Module *Module `gorm:"foreignKey:ModuleID;references:ModuleID" json:"module,omitempty"`
}

func (Permission) TableName() string { return "Permission_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type PermissionRepository interface {
	FindAll(moduleID string) ([]Permission, error)
	FindByID(id string) (*Permission, error)
	Create(p *Permission) error
	Update(p *Permission) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type PermissionUseCase interface {
	GetAll(moduleID string) ([]Permission, error)
	GetByID(id string) (*Permission, error)
	Create(p *Permission) error
	Update(p *Permission) error
	Delete(id string) error
}

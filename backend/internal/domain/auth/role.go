package auth

import "time"

// Role represents the Role_Master table.
type Role struct {
	RoleID          string    `gorm:"column:RoleID;primaryKey" json:"role_id"`
	RoleName        string    `gorm:"column:RoleName;not null" json:"role_name"`
	RoleDescription string    `gorm:"column:RoleDescription;size:255" json:"role_description"`
	RoleCreatedAt   time.Time `gorm:"column:RoleCreatedAt;autoCreateTime" json:"created_at"`
	RoleUpdatedAt   time.Time `gorm:"column:RoleUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Permissions []RolePermission `gorm:"foreignKey:RoleID" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "Role_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type RoleRepository interface {
	FindAll(page, limit int, search string) ([]Role, int64, error)
	FindByID(id string) (*Role, error)
	Create(r *Role) error
	Update(r *Role) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type RoleUseCase interface {
	GetAll(page, limit int, search string) ([]Role, int64, error)
	GetByID(id string) (*Role, error)
	Create(r *Role) error
	Update(r *Role) error
	Delete(id string) error
}

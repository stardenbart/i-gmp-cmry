package auth

import "time"

// Module represents the Module_Master table.
// Modules are the menu/feature groups of the application.
type Module struct {
	ModuleID        string    `gorm:"column:ModuleID;primaryKey" json:"module_id"`
	ModuleName      string    `gorm:"column:ModuleName;not null" json:"module_name"`
	ModuleCreatedAt time.Time `gorm:"column:ModuleCreatedAt;autoCreateTime" json:"created_at"`
	ModuleUpdatedAt time.Time `gorm:"column:ModuleUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Permissions []Permission `gorm:"foreignKey:ModuleID" json:"permissions,omitempty"`
}

func (Module) TableName() string { return "Module_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type ModuleRepository interface {
	FindAll() ([]Module, error)
	FindByID(id string) (*Module, error)
	Create(m *Module) error
	Update(m *Module) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type ModuleUseCase interface {
	GetAll() ([]Module, error)
	GetByID(id string) (*Module, error)
	Create(m *Module) error
	Update(m *Module) error
	Delete(id string) error
}

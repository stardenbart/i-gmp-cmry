package master

import "time"

// Department represents the Department_Master table.
type Department struct {
	DepartmentID        string    `gorm:"column:DepartmentID;primaryKey" json:"department_id"`
	DepartmentName      string    `gorm:"column:DepartmentName;not null" json:"department_name"`
	DepartmentCreatedAt time.Time `gorm:"column:DepartmentCreatedAt;autoCreateTime" json:"created_at"`
	DepartmentUpdatedAt time.Time `gorm:"column:DepartmentUpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (Department) TableName() string { return "Department_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type DepartmentRepository interface {
	FindAll(page, limit int, search string) ([]Department, int64, error)
	FindByID(id string) (*Department, error)
	Create(dept *Department) error
	Update(dept *Department) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type DepartmentUseCase interface {
	GetAll(page, limit int, search string) ([]Department, int64, error)
	GetByID(id string) (*Department, error)
	Create(dept *Department) error
	Update(dept *Department) error
	Delete(id string) error
}

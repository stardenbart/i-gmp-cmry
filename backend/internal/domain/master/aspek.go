package master

import "time"

// Aspek represents the Aspek_Master table.
type Aspek struct {
	AspekID        string    `gorm:"column:AspekID;primaryKey" json:"aspek_id"`
	AreaID         string    `gorm:"column:AreaID;not null" json:"area_id"`
	AspekName      string    `gorm:"column:AspekName;not null" json:"aspek_name"`
	AspekCreatedAt time.Time `gorm:"column:AspekCreatedAt;autoCreateTime" json:"created_at"`
	AspekUpdatedAt time.Time `gorm:"column:AspekUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Area    *Area    `gorm:"foreignKey:AreaID;references:AreaID" json:"area,omitempty"`
	Details []Detail `gorm:"foreignKey:AspekID" json:"details,omitempty"`
}

func (Aspek) TableName() string { return "Aspek_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type AspekRepository interface {
	FindAll(page, limit int, plantID, areaID, search string) ([]Aspek, int64, error)
	FindByID(id string) (*Aspek, error)
	FindByAreaID(areaID string) ([]Aspek, error)
	Create(a *Aspek) error
	Update(a *Aspek) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type AspekUseCase interface {
	GetAll(page, limit int, plantID, areaID, search string) ([]Aspek, int64, error)
	GetByID(id string) (*Aspek, error)
	GetByAreaID(areaID string) ([]Aspek, error)
	Create(a *Aspek) error
	Update(a *Aspek) error
	Delete(id string) error
}

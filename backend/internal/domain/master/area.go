package master

import "time"

// Area represents the Area_Master table.
type Area struct {
	AreaID        string    `gorm:"column:AreaID;primaryKey" json:"area_id"`
	PlantID       *string   `gorm:"column:PlantID" json:"plant_id,omitempty"`
	AreaName      string    `gorm:"column:AreaName;not null" json:"area_name"`
	AreaCreatedAt time.Time `gorm:"column:AreaCreatedAt;autoCreateTime" json:"created_at"`
	AreaUpdatedAt time.Time `gorm:"column:AreaUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Plant    *Plant    `gorm:"foreignKey:PlantID;references:PlantID" json:"plant,omitempty"`
	Kawasans []Kawasan `gorm:"foreignKey:AreaID" json:"kawasans,omitempty"`
	Aspeks   []Aspek   `gorm:"foreignKey:AreaID" json:"aspeks,omitempty"`
}

func (Area) TableName() string { return "Area_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type AreaRepository interface {
	FindAll(page, limit int, plantID, search string) ([]Area, int64, error)
	FindByID(id string) (*Area, error)
	Create(area *Area) error
	Update(area *Area) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type AreaUseCase interface {
	GetAll(page, limit int, plantID, search string) ([]Area, int64, error)
	GetByID(id string) (*Area, error)
	Create(area *Area) error
	Update(area *Area) error
	Delete(id string) error
}

package master

import "time"

// Detail represents the Detail_Master table.
type Detail struct {
	DetailID        string    `gorm:"column:DetailID;primaryKey" json:"detail_id"`
	AspekID         string    `gorm:"column:AspekID;not null" json:"aspek_id"`
	DetailName      string    `gorm:"column:DetailName;size:150;not null" json:"detail_name"`
	DetailCreatedAt time.Time `gorm:"column:DetailCreatedAt;autoCreateTime" json:"created_at"`
	DetailUpdatedAt time.Time `gorm:"column:DetailUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Aspek  *Aspek   `gorm:"foreignKey:AspekID;references:AspekID" json:"aspek,omitempty"`
	Urains []Uraian `gorm:"foreignKey:DetailID" json:"urains,omitempty"`
}

func (Detail) TableName() string { return "Detail_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type DetailRepository interface {
	FindAll(page, limit int, aspekID, search string) ([]Detail, int64, error)
	FindByID(id string) (*Detail, error)
	FindByAspekID(aspekID string) ([]Detail, error)
	Create(d *Detail) error
	Update(d *Detail) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type DetailUseCase interface {
	GetAll(page, limit int, aspekID, search string) ([]Detail, int64, error)
	GetByID(id string) (*Detail, error)
	GetByAspekID(aspekID string) ([]Detail, error)
	Create(d *Detail) error
	Update(d *Detail) error
	Delete(id string) error
}

package master

import "time"

// Kawasan represents the Kawasan_Master table.
type Kawasan struct {
	KawasanID        string    `gorm:"column:KawasanID;primaryKey" json:"kawasan_id"`
	AreaID           string    `gorm:"column:AreaID;not null" json:"area_id"`
	KawasanName      string    `gorm:"column:KawasanName;not null" json:"kawasan_name"`
	KawasanCreatedAt time.Time `gorm:"column:KawasanCreatedAt;autoCreateTime" json:"created_at"`
	KawasanUpdatedAt time.Time `gorm:"column:KawasanUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Area           *Area           `gorm:"foreignKey:AreaID;references:AreaID" json:"area,omitempty"`
	DetailKawasans []DetailKawasan `gorm:"foreignKey:KawasanID" json:"detail_kawasans,omitempty"`
}

func (Kawasan) TableName() string { return "Kawasan_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type KawasanRepository interface {
	FindAll(page, limit int, areaID, search string) ([]Kawasan, int64, error)
	FindByID(id string) (*Kawasan, error)
	FindByAreaID(areaID string) ([]Kawasan, error)
	Create(k *Kawasan) error
	Update(k *Kawasan) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type KawasanUseCase interface {
	GetAll(page, limit int, areaID, search string) ([]Kawasan, int64, error)
	GetByID(id string) (*Kawasan, error)
	GetByAreaID(areaID string) ([]Kawasan, error)
	Create(k *Kawasan) error
	Update(k *Kawasan) error
	Delete(id string) error
}

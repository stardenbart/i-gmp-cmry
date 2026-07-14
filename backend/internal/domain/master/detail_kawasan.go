package master

import "time"

// DetailKawasan represents the DetailKawasan_Master table.
type DetailKawasan struct {
	DetailKawasanID        string    `gorm:"column:DetailKawasanID;primaryKey" json:"detail_kawasan_id"`
	KawasanID              string    `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	DetailKawasanName      string    `gorm:"column:DetailKawasanName;not null" json:"detail_kawasan_name"`
	DetailKawasanCreatedAt time.Time `gorm:"column:DetailKawasanCreatedAt;autoCreateTime" json:"created_at"`
	DetailKawasanUpdatedAt time.Time `gorm:"column:DetailKawasanUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Kawasan *Kawasan `gorm:"foreignKey:KawasanID;references:KawasanID" json:"kawasan,omitempty"`
}

func (DetailKawasan) TableName() string { return "DetailKawasan_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type DetailKawasanRepository interface {
	FindAll(page, limit int, kawasanID, search string) ([]DetailKawasan, int64, error)
	FindByID(id string) (*DetailKawasan, error)
	FindByKawasanID(kawasanID string) ([]DetailKawasan, error)
	Create(dk *DetailKawasan) error
	Update(dk *DetailKawasan) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type DetailKawasanUseCase interface {
	GetAll(page, limit int, kawasanID, search string) ([]DetailKawasan, int64, error)
	GetByID(id string) (*DetailKawasan, error)
	GetByKawasanID(kawasanID string) ([]DetailKawasan, error)
	Create(dk *DetailKawasan) error
	Update(dk *DetailKawasan) error
	Delete(id string) error
}

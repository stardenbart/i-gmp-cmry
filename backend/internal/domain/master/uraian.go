package master

import "time"

// Uraian represents the Uraian_Master table.
// UraianText is the checklist item text, StandardScore is the base score.
type Uraian struct {
	UraianID        string    `gorm:"column:UraianID;primaryKey" json:"uraian_id"`
	DetailID        string    `gorm:"column:DetailID;not null" json:"detail_id"`
	UraianText      string    `gorm:"column:UraianText;size:500;not null" json:"uraian_text"`
	StandardScore   int       `gorm:"column:StandardScore;not null;default:2" json:"standard_score"`
	UraianCreatedAt time.Time `gorm:"column:UraianCreatedAt;autoCreateTime" json:"created_at"`
	UraianUpdatedAt time.Time `gorm:"column:UraianUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Detail *Detail `gorm:"foreignKey:DetailID;references:DetailID" json:"detail,omitempty"`
}

func (Uraian) TableName() string { return "Uraian_Master" }

// ─── Repository Interface ──────────────────────────────────────────────────

type UraianRepository interface {
	FindAll(page, limit int, plantID, detailID, search string) ([]Uraian, int64, error)
	FindByID(id string) (*Uraian, error)
	FindByDetailID(detailID string) ([]Uraian, error)
	Create(u *Uraian) error
	Update(u *Uraian) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type UraianUseCase interface {
	GetAll(page, limit int, plantID, detailID, search string) ([]Uraian, int64, error)
	GetByID(id string) (*Uraian, error)
	GetByDetailID(detailID string) ([]Uraian, error)
	Create(u *Uraian) error
	Update(u *Uraian) error
	Delete(id string) error
}

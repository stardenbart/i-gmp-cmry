package pic

import "time"

// PICMapping represents the PIC_Mapping table.
// Maps a user as Person-in-Charge for a specific Area/Kawasan.
type PICMapping struct {
	PICMapID           string    `gorm:"column:PICMapID;primaryKey" json:"pic_map_id"`
	AreaID             string    `gorm:"column:AreaID;not null" json:"area_id"`
	KawasanID          string    `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	UserID             string    `gorm:"column:UserID;not null" json:"user_id"`
	KategoriPIC        string    `gorm:"column:KategoriPIC;size:50" json:"kategori_pic"`
	PICMappingCreatedAt time.Time `gorm:"column:PICMappingCreatedAt;autoCreateTime" json:"created_at"`
	PICMappingUpdatedAt time.Time `gorm:"column:PICMappingUpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (PICMapping) TableName() string { return "PIC_Mapping" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

type CreatePICMappingRequest struct {
	AreaID      string `json:"area_id" validate:"required"`
	KawasanID   string `json:"kawasan_id" validate:"required"`
	UserID      string `json:"user_id" validate:"required"`
	KategoriPIC string `json:"kategori_pic"`
}

type UpdatePICMappingRequest struct {
	KategoriPIC string `json:"kategori_pic"`
	UserID      string `json:"user_id"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type PICMappingRepository interface {
	FindAll(page, limit int, areaID, kawasanID string) ([]PICMapping, int64, error)
	FindByID(id string) (*PICMapping, error)
	FindByUserID(userID string) ([]PICMapping, error)
	FindByAreaAndKawasan(areaID, kawasanID string) ([]PICMapping, error)
	Create(p *PICMapping) error
	Update(p *PICMapping) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type PICMappingUseCase interface {
	GetAll(page, limit int, areaID, kawasanID string) ([]PICMapping, int64, error)
	GetByID(id string) (*PICMapping, error)
	GetByUserID(userID string) ([]PICMapping, error)
	Create(req *CreatePICMappingRequest) (*PICMapping, error)
	Update(id string, req *UpdatePICMappingRequest) (*PICMapping, error)
	Delete(id string) error
}

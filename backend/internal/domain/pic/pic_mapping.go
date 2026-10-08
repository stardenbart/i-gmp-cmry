package pic

import "time"

// PICMapping represents the PIC_Mapping table.
// Maps a user as Person-in-Charge for a specific Area/Kawasan.
type PICMapping struct {
	PICMapID            string    `gorm:"column:PICMapID;primaryKey" json:"pic_map_id"`
	AreaID              *string   `gorm:"column:AreaID" json:"area_id,omitempty"`
	KawasanID           string    `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	UserID              string    `gorm:"column:UserID;not null" json:"user_id"`
	KategoriPIC         string    `gorm:"column:KategoriPIC;size:50" json:"kategori_pic"`
	PICMappingCreatedAt time.Time `gorm:"column:PICMappingCreatedAt;autoCreateTime" json:"created_at"`
	PICMappingUpdatedAt time.Time `gorm:"column:PICMappingUpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (PICMapping) TableName() string { return "PIC_Mapping" }

// ResponsibleUser is the minimal user info needed to notify/email someone
// resolved as responsible for a Kawasan (PIC, Manager, Supervisor, ...).
// Deliberately not authdomain.User: that package already imports pic (for
// User.PICMappings), so reusing it here would create an import cycle.
type ResponsibleUser struct {
	UserID   string
	FullName string
	Email    string
	PlantID  *string
	// KategoriPIC is the mapping role (Manager, Supervisor, Staff, ...).
	KategoriPIC string
}

// ─── DTOs ──────────────────────────────────────────────────────────────────

type CreatePICMappingRequest struct {
	AreaID      *string `json:"area_id"`
	KawasanID   string  `json:"kawasan_id" validate:"required"`
	UserID      string  `json:"user_id" validate:"required"`
	KategoriPIC string  `json:"kategori_pic"`
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
	// FindResponsibleUsers returns every user PIC-mapped to kawasanID,
	// optionally filtered to one KategoriPIC (e.g. "Manager"); pass "" to
	// get all categories (PIC/Manager/Supervisor/Staff alike).
	FindResponsibleUsers(kawasanID string, kategori string) ([]ResponsibleUser, error)
	// FindResponsibleUsersByResultID resolves the Kawasan an Inspection_Result
	// belongs to (via its Inspection_Header) and returns its responsible
	// users the same way as FindResponsibleUsers. Lets callers that only
	// have a ResultID (e.g. Issue creation) reach the right Kawasan without
	// needing their own InspectionHeaderRepository dependency.
	FindResponsibleUsersByResultID(resultID string, kategori string) ([]ResponsibleUser, error)
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

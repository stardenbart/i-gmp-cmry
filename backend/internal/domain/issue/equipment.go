package issue

import (
	"time"
)

// Equipment represents the Equipment_Master table.
type Equipment struct {
	EquipmentID        string    `gorm:"column:EquipmentID;primaryKey" json:"equipment_id"`
	KawasanID          string    `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	EquipmentCode      string    `gorm:"column:EquipmentCode;unique" json:"equipment_code,omitempty"`
	EquipmentName      string    `gorm:"column:EquipmentName;not null" json:"equipment_name"`
	EquipmentType      string    `gorm:"column:EquipmentType" json:"equipment_type,omitempty"`
	EquipmentStatus    string    `gorm:"column:EquipmentStatus;not null;default:Active" json:"equipment_status"`
	EquipmentCreatedAt time.Time `gorm:"column:EquipmentCreatedAt;autoCreateTime" json:"created_at"`
	EquipmentUpdatedAt time.Time `gorm:"column:EquipmentUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Joined/Preloaded
	KawasanName string `gorm:"column:KawasanName;->" json:"kawasan_name,omitempty"`
}

func (Equipment) TableName() string { return "Equipment_Master" }

// DTOs
type CreateEquipmentRequest struct {
	KawasanID       string `json:"kawasan_id" validate:"required"`
	EquipmentCode   string `json:"equipment_code"`
	EquipmentName   string `json:"equipment_name" validate:"required,max=150"`
	EquipmentType   string `json:"equipment_type"`
	EquipmentStatus string `json:"equipment_status"`
}

type UpdateEquipmentRequest struct {
	KawasanID       string `json:"kawasan_id"`
	EquipmentCode   string `json:"equipment_code"`
	EquipmentName   string `json:"equipment_name" validate:"omitempty,max=150"`
	EquipmentType   string `json:"equipment_type"`
	EquipmentStatus string `json:"equipment_status"`
}

// Interfaces
type EquipmentRepository interface {
	FindAll(page, limit int, kawasanID, search string) ([]Equipment, int64, error)
	FindByID(id string) (*Equipment, error)
	Create(eq *Equipment) error
	Update(eq *Equipment) error
	Delete(id string) error
}

type EquipmentUseCase interface {
	GetAll(page, limit int, kawasanID, search string) ([]Equipment, int64, error)
	GetByID(id string) (*Equipment, error)
	Create(req *CreateEquipmentRequest) (*Equipment, error)
	Update(id string, req *UpdateEquipmentRequest) (*Equipment, error)
	Delete(id string) error
}

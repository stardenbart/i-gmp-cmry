package issue

import (
	"time"
)

// Infrastructure represents the Infrastructure_Master table.
type Infrastructure struct {
	InfrastructureID        string    `gorm:"column:InfrastructureID;primaryKey" json:"infrastructure_id"`
	KawasanID               string    `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	InfrastructureCode      string    `gorm:"column:InfrastructureCode;unique" json:"infrastructure_code,omitempty"`
	InfrastructureName      string    `gorm:"column:InfrastructureName;not null" json:"infrastructure_name"`
	InfrastructureType      string    `gorm:"column:InfrastructureType" json:"infrastructure_type,omitempty"`
	InfrastructureStatus    string    `gorm:"column:InfrastructureStatus;not null;default:Active" json:"infrastructure_status"`
	InfrastructureCreatedAt time.Time `gorm:"column:InfrastructureCreatedAt;autoCreateTime" json:"created_at"`
	InfrastructureUpdatedAt time.Time `gorm:"column:InfrastructureUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Joined/Preloaded
	KawasanName string `gorm:"column:KawasanName;->" json:"kawasan_name,omitempty"`
}

func (Infrastructure) TableName() string { return "Infrastructure_Master" }

// DTOs
type CreateInfrastructureRequest struct {
	KawasanID            string `json:"kawasan_id" validate:"required"`
	InfrastructureCode   string `json:"infrastructure_code"`
	InfrastructureName   string `json:"infrastructure_name" validate:"required,max=150"`
	InfrastructureType   string `json:"infrastructure_type"`
	InfrastructureStatus string `json:"infrastructure_status"`
}

type UpdateInfrastructureRequest struct {
	KawasanID            string `json:"kawasan_id"`
	InfrastructureCode   string `json:"infrastructure_code"`
	InfrastructureName   string `json:"infrastructure_name" validate:"omitempty,max=150"`
	InfrastructureType   string `json:"infrastructure_type"`
	InfrastructureStatus string `json:"infrastructure_status"`
}

// Interfaces
type InfrastructureRepository interface {
	FindAll(page, limit int, kawasanID, search string) ([]Infrastructure, int64, error)
	FindByID(id string) (*Infrastructure, error)
	Create(inf *Infrastructure) error
	Update(inf *Infrastructure) error
	Delete(id string) error
}

type InfrastructureUseCase interface {
	GetAll(page, limit int, kawasanID, search string) ([]Infrastructure, int64, error)
	GetByID(id string) (*Infrastructure, error)
	Create(req *CreateInfrastructureRequest) (*Infrastructure, error)
	Update(id string, req *UpdateInfrastructureRequest) (*Infrastructure, error)
	Delete(id string) error
}

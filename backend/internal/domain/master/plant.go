package master

import (
	"time"
)

// Plant represents the Plant_Master table.
type Plant struct {
	PlantID        string    `gorm:"column:PlantID;primaryKey" json:"plant_id"`
	PlantCode      string    `gorm:"column:PlantCode;not null;unique" json:"plant_code"`
	PlantName      string    `gorm:"column:PlantName;not null" json:"plant_name"`
	Address        string    `gorm:"column:Address" json:"address,omitempty"`
	PlantCreatedAt time.Time `gorm:"column:PlantCreatedAt;autoCreateTime" json:"created_at"`
	PlantUpdatedAt time.Time `gorm:"column:PlantUpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (Plant) TableName() string { return "Plant_Master" }

// DTOs
type CreatePlantRequest struct {
	PlantCode string `json:"plant_code" validate:"required,max=20"`
	PlantName string `json:"plant_name" validate:"required,max=100"`
	Address   string `json:"address"`
}

type UpdatePlantRequest struct {
	PlantCode string `json:"plant_code" validate:"omitempty,max=20"`
	PlantName string `json:"plant_name" validate:"omitempty,max=100"`
	Address   string `json:"address"`
}

// Repository & UseCase Interfaces
type PlantRepository interface {
	FindAll(page, limit int, search string) ([]Plant, int64, error)
	FindByID(id string) (*Plant, error)
	FindByCode(code string) (*Plant, error)
	Create(p *Plant) error
	Update(p *Plant) error
	Delete(id string) error
}

type PlantUseCase interface {
	GetAll(page, limit int, search string) ([]Plant, int64, error)
	GetByID(id string) (*Plant, error)
	Create(req *CreatePlantRequest) (*Plant, error)
	Update(id string, req *UpdatePlantRequest) (*Plant, error)
	Delete(id string) error
}

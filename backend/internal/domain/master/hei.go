package master

import (
	"context"
	"time"
)

// HEIMaster represents the HEI_Master table for dynamic HEI framework.
type HEIMaster struct {
	HEIID        string    `gorm:"column:HEIID;primaryKey" json:"hei_id"`
	CategoryName string    `gorm:"column:CategoryName;not null" json:"category_name"`
	HEICode      string    `gorm:"column:HEICode" json:"hei_code"`
	HEIName      string    `gorm:"column:HEIName;not null" json:"hei_name"`
	Description  string    `gorm:"column:Description" json:"description"`
	Status       string    `gorm:"column:Status;default:Active" json:"status"`
	CreatedAt    time.Time `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:UpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (HEIMaster) TableName() string { return "HEI_Master" }

type CreateHEIRequest struct {
	CategoryName string `json:"category_name" validate:"required"`
	HEICode      string `json:"hei_code"`
	HEIName      string `json:"hei_name" validate:"required"`
	Description  string `json:"description"`
	Status       string `json:"status"`
}

type UpdateHEIRequest struct {
	CategoryName string `json:"category_name"`
	HEICode      string `json:"hei_code"`
	HEIName      string `json:"hei_name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
}

type HEIRepository interface {
	FindAll(page, limit int, category, search string) ([]HEIMaster, int64, error)
	FindByID(id string) (*HEIMaster, error)
	FindCategories() ([]string, error)
	Create(h *HEIMaster) error
	Update(h *HEIMaster) error
	Delete(id string) error
}

type HEIUseCase interface {
	GetAll(page, limit int, category, search string) ([]HEIMaster, int64, error)
	GetByID(id string) (*HEIMaster, error)
	GetCategories(ctx context.Context) ([]string, error)
	Create(req *CreateHEIRequest) (*HEIMaster, error)
	Update(id string, req *UpdateHEIRequest) (*HEIMaster, error)
	Delete(id string) error
}

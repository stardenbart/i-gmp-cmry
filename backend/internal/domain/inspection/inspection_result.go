package inspection

import "time"

// InspectionResult represents the Inspection_Result table.
// Stores the actual checklist scoring per Uraian item.
type InspectionResult struct {
	ResultID                  string    `gorm:"column:ResultID;primaryKey" json:"result_id"`
	InspectionID              string    `gorm:"column:InspectionID;not null" json:"inspection_id"`
	UraianID                  string    `gorm:"column:UraianID;not null" json:"uraian_id"`
	Checking                  string    `gorm:"column:Checking;size:20" json:"checking"`    // OK, NG, NA
	Nilai                     int       `gorm:"column:Nilai;default:0" json:"nilai"`
	Keterangan                string    `gorm:"column:Keterangan;size:255" json:"keterangan"`
	InspectionResultCreatedAt time.Time `gorm:"column:InspectionResultCreatedAt;autoCreateTime" json:"created_at"`
	InspectionResultUpdatedAt time.Time `gorm:"column:InspectionResultUpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (InspectionResult) TableName() string { return "Inspection_Result" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

type SaveResultRequest struct {
	UraianID   string `json:"uraian_id" validate:"required"`
	Checking   string `json:"checking" validate:"required,oneof=OK NG NA"`
	Nilai      int    `json:"nilai" validate:"min=0"`
	Keterangan string `json:"keterangan"`
}

type BulkSaveResultRequest struct {
	InspectionID string              `json:"inspection_id" validate:"required"`
	Results      []SaveResultRequest `json:"results" validate:"required,dive"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type InspectionResultRepository interface {
	FindByInspectionID(inspectionID string) ([]InspectionResult, error)
	FindByID(id string) (*InspectionResult, error)
	Create(r *InspectionResult) error
	BulkCreate(rs []InspectionResult) error
	Update(r *InspectionResult) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type InspectionResultUseCase interface {
	GetByInspectionID(inspectionID string) ([]InspectionResult, error)
	GetByID(id string) (*InspectionResult, error)
	BulkSave(req *BulkSaveResultRequest) error
	Update(id string, req *SaveResultRequest) (*InspectionResult, error)
	Delete(id string) error
}

package inspection

import "time"

// InspectionStatus defines allowed values for the inspection header status.
type InspectionStatus string

const (
	InspectionStatusDraft     InspectionStatus = "Draft"
	InspectionStatusOngoing   InspectionStatus = "Ongoing"
	InspectionStatusCompleted InspectionStatus = "Completed"
	InspectionStatusApproved  InspectionStatus = "Approved"
)

// InspectionHeader represents the Inspection_Header table.
type InspectionHeader struct {
	InspectionID              string           `gorm:"column:InspectionID;primaryKey" json:"inspection_id"`
	AreaID                    string           `gorm:"column:AreaID;not null" json:"area_id"`
	KawasanID                 string           `gorm:"column:KawasanID;not null" json:"kawasan_id"`
	DetailKawasanID           string           `gorm:"column:DetailKawasanID;not null" json:"detail_kawasan_id"`
	InspectorID               string           `gorm:"column:InspectorID;not null" json:"inspector_id"`
	InspectionHeaderStatus    InspectionStatus `gorm:"column:InspectionHeaderStatus;not null;default:Draft" json:"status"`
	InspectionHeaderCreatedAt time.Time        `gorm:"column:InspectionHeaderCreatedAt;autoCreateTime" json:"created_at"`
	InspectionheaderUpdatedAt time.Time        `gorm:"column:InspectionheaderUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Results []InspectionResult `gorm:"foreignKey:InspectionID" json:"results,omitempty"`
}

func (InspectionHeader) TableName() string { return "Inspection_Header" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

type CreateInspectionRequest struct {
	AreaID          string `json:"area_id" validate:"required"`
	KawasanID       string `json:"kawasan_id" validate:"required"`
	DetailKawasanID string `json:"detail_kawasan_id" validate:"required"`
}

type UpdateInspectionStatusRequest struct {
	Status InspectionStatus `json:"status" validate:"required,oneof=Draft Ongoing Completed Approved"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type InspectionHeaderRepository interface {
	FindAll(page, limit int, areaID, status, inspectorID string) ([]InspectionHeader, int64, error)
	FindByID(id string) (*InspectionHeader, error)
	Create(h *InspectionHeader) error
	Update(h *InspectionHeader) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type InspectionHeaderUseCase interface {
	GetAll(page, limit int, areaID, status, inspectorID string) ([]InspectionHeader, int64, error)
	GetByID(id string) (*InspectionHeader, error)
	Create(inspectorID string, req *CreateInspectionRequest) (*InspectionHeader, error)
	UpdateStatus(id string, actorID string, req *UpdateInspectionStatusRequest) (*InspectionHeader, error)
	Delete(id string) error
}

package inspection

import (
	"context"
	"time"
)

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
	InspectionHeaderUpdatedAt time.Time        `gorm:"column:InspectionHeaderUpdatedAt;autoUpdateTime" json:"updated_at"`

	SessionID *string    `gorm:"column:SessionID;size:255" json:"session_id"`
	LockedAt  *time.Time `gorm:"column:LockedAt" json:"locked_at"`

	// Joined Name Fields (not saved to DB)
	AreaName          string   `gorm:"column:AreaName;->" json:"area_name,omitempty"`
	KawasanName       string   `gorm:"column:KawasanName;->" json:"kawasan_name,omitempty"`
	DetailKawasanName string   `gorm:"column:DetailKawasanName;->" json:"detail_kawasan_name,omitempty"`
	InspectorName     string   `gorm:"column:InspectorName;->" json:"inspector_name,omitempty"`
	Score             *float64 `gorm:"column:Score;->" json:"score,omitempty"`

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

type TrendData struct {
	Month string `json:"month"`
	Rate  int    `json:"rate"` // using rate to match frontend Recharts generic mapping, or count
}

// ─── Repository Interface ──────────────────────────────────────────────────

type InspectionHeaderRepository interface {
	FindAll(page, limit int, plantID, areaID, status, inspectorID string) ([]InspectionHeader, int64, error)
	FindByID(id string) (*InspectionHeader, error)
	FindByIDWithCtx(ctx context.Context, id string) (*InspectionHeader, error)
	FindActiveByKawasan(kawasanID string) ([]InspectionHeader, error)
	FindActiveByDetailKawasan(detailKawasanID string) ([]InspectionHeader, error)
	FindActiveByInspector(inspectorID string) ([]InspectionHeader, error)
	FindActiveByInspectorAndDetailKawasan(inspectorID, detailKawasanID string) ([]InspectionHeader, error)
	CountCompletedThisMonthByDetailKawasan(detailKawasanID string, year int, month int) (int64, error)
	CountCompletedByAreaAndDetailKawasan(areaID, detailKawasanID string) (int64, error)
	GetTrendByContext(contextID string, year int) ([]TrendData, error)
	GetFullChecklist(areaID, inspectionID string) (*FullChecklist, error)
	GetFullChecklistWithCtx(ctx context.Context, areaID, inspectionID string) (*FullChecklist, error)
	Create(h *InspectionHeader) error
	Update(h *InspectionHeader) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type InspectionHeaderUseCase interface {
	GetAll(page, limit int, plantID, areaID, status, inspectorID string) ([]InspectionHeader, int64, error)
	GetByID(id string) (*InspectionHeader, error)
	GetAreaStatus(areaID string) (AreaProgress, error)
	GetTrend(contextID string, year int) ([]TrendData, error)
	GetChecklist(id string) (*FullChecklist, error)
	Create(inspectorID string, req *CreateInspectionRequest) (*InspectionHeader, error)
	UpdateStatus(id string, actorID string, req *UpdateInspectionStatusRequest) (*InspectionHeader, error)
	Delete(id string) error
}

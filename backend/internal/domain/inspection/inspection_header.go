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
	Month       string `json:"month"`
	Date        string `json:"date"`
	Rate        int    `json:"rate"` // using rate to match frontend Recharts generic mapping, or count
	TotalIssues int    `json:"total_issues"`
}

// InspectionPeriodInfo describes the currently-running "inspection period"
// (the recurring monthly cycle a DetailKawasan must be re-inspected
// within) for a given Area's plant, resolved from the admin-configurable
// INSPECTION_PERIOD_CUTOFF_DAY setting. See
// inspectionusecase.ResolveInspectionPeriod for how the boundaries are
// computed.
type InspectionPeriodInfo struct {
	CutoffDay   int       `json:"cutoff_day"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
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
	// CountCompletedInPeriod counts Completed/Approved inspections for a
	// DetailKawasan whose InspectionHeaderCreatedAt falls within [start, end)
	// — the boundaries of one "inspection period" (see
	// inspectionusecase.ResolveInspectionPeriod). Replaces the old
	// calendar-month-only CountCompletedThisMonthByDetailKawasan and the
	// never-time-filtered CountCompletedByAreaAndDetailKawasan, unifying
	// Kawasan- and Area-level completion checks onto the same admin
	// configurable cutoff.
	CountCompletedInPeriod(detailKawasanID string, start, end time.Time) (int64, error)
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
	// GetByIDScoped is GetByID plus a per-record plant check: userPlantID
	// ("" for Super Admin / global-scope users) must match the plant that
	// owns the inspection's AreaID, via authz.PlantMatches. Returns the same
	// "inspection not found" error as GetByID for a denied scope — a
	// wrong-plant caller must not be able to distinguish "doesn't exist"
	// from "exists but isn't yours" by ID alone (GetByID/GetAll never
	// enforced this on their own; a guessable, unauthenticated-by-plant ID
	// was the only thing stopping cross-plant reads before this).
	GetByIDScoped(id, userPlantID string) (*InspectionHeader, error)
	GetAreaStatus(areaID string) (AreaProgress, error)
	GetCurrentPeriodInfo(areaID string) (InspectionPeriodInfo, error)
	GetTrend(contextID string, year int) ([]TrendData, error)
	GetChecklist(id string) (*FullChecklist, error)
	Create(inspectorID string, req *CreateInspectionRequest) (*InspectionHeader, error)
	UpdateStatus(id string, actorID string, req *UpdateInspectionStatusRequest) (*InspectionHeader, error)
	Delete(id string) error
}

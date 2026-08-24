package issue

import (
	"time"

	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
)

// IssueStatus defines allowed values for issue status.
type IssueStatus string

const (
	IssueStatusOpen              IssueStatus = "Open"
	IssueStatusInProgress        IssueStatus = "InProgress"
	IssueStatusPendingValidation IssueStatus = "PendingValidation"
	IssueStatusClosed            IssueStatus = "Closed"
	IssueStatusVerified          IssueStatus = "Verified"
	IssueStatusOpenOverdue       IssueStatus = "OpenOverdue"
	IssueStatusClosedOverdue     IssueStatus = "ClosedOverdue"
)

// WOWRStatus defines allowed values for WOWR validation status.
type WOWRStatus string

const (
	WOWRStatusNone              WOWRStatus = "None"
	WOWRStatusPendingValidation WOWRStatus = "PendingValidation"
	WOWRStatusVerified          WOWRStatus = "Verified"
	WOWRStatusRejected          WOWRStatus = "Rejected"
)

// IssueHEI represents the Issue_HEI junction table (1-to-1 per Issue).
type IssueHEI struct {
	IssueHEIID string    `gorm:"column:IssueHEIID;primaryKey" json:"issue_hei_id"`
	IssueID    string    `gorm:"column:IssueID;unique;not null" json:"issue_id"`
	HEIID      *string   `gorm:"column:HEIID" json:"hei_id,omitempty"`
	CreatedAt  time.Time `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:UpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	HEI *masterdomain.HEIMaster `gorm:"foreignKey:HEIID;references:HEIID" json:"hei,omitempty"`
}

func (IssueHEI) TableName() string { return "Issue_HEI" }

// Issue represents the Issue table.
type Issue struct {
	DetailKawasanID  string      `gorm:"column:DetailKawasanID" json:"detail_kawasan_id"`
	IssueID             string        `gorm:"column:IssueID;primaryKey" json:"issue_id"`
	ResultID            string        `gorm:"column:ResultID;not null" json:"result_id"`
	IssuePICUserID      string        `gorm:"column:IssuePICUserID;not null" json:"issue_pic_user_id"`
	DueDate             *time.Time    `gorm:"column:DueDate" json:"due_date"`
	IssueStatus         IssueStatus   `gorm:"column:IssueStatus;not null;default:Open" json:"issue_status"`
	ComputedIssueStatus IssueStatus   `gorm:"-" json:"computed_status,omitempty"`
	FollowUpDelay       *int          `gorm:"column:FollowUpDelay" json:"follow_up_delay,omitempty"`
	Label               string        `gorm:"column:Label;size:100" json:"label"`
	NeedsWOWR           bool          `gorm:"column:NeedsWOWR;default:false" json:"needs_wo_wr"`
	WO_ID               string        `gorm:"column:WO_ID;size:100" json:"wo_id"`
	WR_ID               string        `gorm:"column:WR_ID;size:100" json:"wr_id"`
	WOWRStatus          WOWRStatus    `gorm:"column:WOWRStatus;default:None" json:"wowr_status"`
	Keterangan          string        `gorm:"column:Keterangan;size:255" json:"keterangan"`
	PICName             string        `gorm:"-" json:"pic_name"`
	IssueCreatedAt      time.Time     `gorm:"column:IssueCreatedAt;autoCreateTime" json:"created_at"`
	IssueUpdatedAt      time.Time     `gorm:"column:IssueUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Joined Name Fields (not saved to DB)
	AreaName           string `gorm:"column:AreaName;->" json:"area_name,omitempty"`
	KawasanName        string `gorm:"column:KawasanName;->" json:"kawasan_name,omitempty"`
	DetailKawasanName  string `gorm:"column:DetailKawasanName;->" json:"detail_kawasan_name,omitempty"`
	AspekName          string `gorm:"column:AspekName;->" json:"aspek_name,omitempty"`
	DetailAspekName    string `gorm:"column:DetailAspekName;->" json:"detail_aspek_name,omitempty"`
	UraianText         string `gorm:"column:UraianText;->" json:"uraian_text,omitempty"`
	HEICategory        string `gorm:"column:HEICategory;->" json:"hei_category,omitempty"`
	HEIName            string `gorm:"column:HEIName;->" json:"hei_name,omitempty"`
	HabitName          string `gorm:"column:HabitName;->" json:"habit_name,omitempty"`
	EquipmentName      string `gorm:"column:EquipmentName;->" json:"equipment_name,omitempty"`
	InfrastructureName string `gorm:"column:InfrastructureName;->" json:"infrastructure_name,omitempty"`

	// Relations
	Photos []IssuePhoto `gorm:"foreignKey:IssueID" json:"photos,omitempty"`
	HEI    *IssueHEI    `gorm:"foreignKey:IssueID;references:IssueID" json:"hei,omitempty"`
}

func (Issue) TableName() string { return "Issue" }

func (i *Issue) ComputedStatus(now time.Time) IssueStatus {
	isOverdue := i.DueDate != nil && now.After(*i.DueDate)
	switch i.IssueStatus {
	case IssueStatusOpen, IssueStatusInProgress:
		if isOverdue {
			return IssueStatusOpenOverdue
		}
		return i.IssueStatus
	case IssueStatusClosed, IssueStatusVerified:
		if (i.FollowUpDelay != nil && *i.FollowUpDelay > 0) || isOverdue {
			return IssueStatusClosedOverdue
		}
		return i.IssueStatus
	default:
		return i.IssueStatus
	}
}

// ─── DTOs ──────────────────────────────────────────────────────────────────

type CreateIssueRequest struct {
	ResultID         string     `json:"result_id" validate:"required"`
	IssuePICUserID   string     `json:"issue_pic_user_id" validate:"required"`
	DueDate          *time.Time `json:"due_date"`
	Label            string     `json:"label"`
	NeedsWOWR        bool       `json:"needs_wo_wr"`
	WO_ID            string     `json:"wo_id"`
	WR_ID            string     `json:"wr_id"`
	WOWRStatus       WOWRStatus `json:"wowr_status"`
	HabitID          *string    `json:"habit_id"`
	EquipmentID      *string    `json:"equipment_id"`
	InfrastructureID *string    `json:"infrastructure_id"`
	Keterangan       string     `json:"keterangan"`
}

type UpdateIssueRequest struct {
	IssuePICUserID   string      `json:"issue_pic_user_id"`
	DueDate          *time.Time  `json:"due_date"`
	IssueStatus      IssueStatus `json:"issue_status" validate:"omitempty,oneof=Open InProgress PendingValidation Closed Verified"`
	Label            string      `json:"label"`
	NeedsWOWR        *bool       `json:"needs_wo_wr"`
	WO_ID            string      `json:"wo_id"`
	WR_ID            string      `json:"wr_id"`
	WOWRStatus       WOWRStatus  `json:"wowr_status" validate:"omitempty,oneof=None PendingValidation Verified Rejected"`
	HabitID          *string     `json:"habit_id"`
	EquipmentID      *string     `json:"equipment_id"`
	InfrastructureID *string     `json:"infrastructure_id"`
	Keterangan       string      `json:"keterangan"`
}

// ─── Repository Interfaces ─────────────────────────────────────────────────

type IssueRepository interface {
	FindAll(page, limit int, plantID, status, picUserID string, needsWOWR *bool) ([]Issue, int64, error)
	FindByID(id string) (*Issue, error)
	FindByResultID(resultID string) (*Issue, error)
	FindActiveByUraianAndDetailKawasan(uraianID, detailKawasanID string) (*Issue, error)
	FindActiveByResultContext(resultID string) (*Issue, error)
	ConsolidateDuplicateActiveIssues() error
	Create(i *Issue) error
	Update(i *Issue) error
	Delete(id string) error
}

type IssueHEIRepository interface {
	UpsertByIssueID(issueID string, hei *IssueHEI) error
	FindByIssueID(issueID string) (*IssueHEI, error)
	DeleteByIssueID(issueID string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type IssueUseCase interface {
	GetAll(page, limit int, plantID, status, picUserID string, needsWOWR *bool) ([]Issue, int64, error)
	GetByID(id string) (*Issue, error)
	GetByResultID(resultID string) (*Issue, error)
	Create(actorID string, req *CreateIssueRequest) (*Issue, error)
	Update(id string, actorID string, req *UpdateIssueRequest) (*Issue, error)
	ExtendDueDate(id string, actorID string, newDueDate time.Time) (*Issue, error)
	Delete(id string, actorID string) error
	CloseByResultID(resultID string, actorID string) error
	ConsolidateDuplicateActiveIssues() error
}

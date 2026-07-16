package issue

import (
	"time"
)

// IssueStatus defines allowed values for issue status.
type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "Open"
	IssueStatusInProgress      IssueStatus = "InProgress"
	IssueStatusPendingValidation IssueStatus = "PendingValidation"
	IssueStatusClosed          IssueStatus = "Closed"
	IssueStatusVerified        IssueStatus = "Verified"
)

// Issue represents the Issue table.
type Issue struct {
	IssueID        string      `gorm:"column:IssueID;primaryKey" json:"issue_id"`
	ResultID       string      `gorm:"column:ResultID;not null" json:"result_id"`
	IssuePICUserID string      `gorm:"column:IssuePICUserID;not null" json:"issue_pic_user_id"`
	DueDate        *time.Time  `gorm:"column:DueDate" json:"due_date"`
	IssueStatus    IssueStatus `gorm:"column:IssueStatus;not null;default:Open" json:"issue_status"`
	Keterangan     string      `gorm:"column:Keterangan;size:255" json:"keterangan"`
	IssueCreatedAt time.Time   `gorm:"column:IssueCreatedAt;autoCreateTime" json:"created_at"`
	IssueUpdatedAt time.Time   `gorm:"column:IssueUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations
	Photos []IssuePhoto `gorm:"foreignKey:IssueID" json:"photos,omitempty"`
}

func (Issue) TableName() string { return "Issue" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

type CreateIssueRequest struct {
	ResultID       string     `json:"result_id" validate:"required"`
	IssuePICUserID string     `json:"issue_pic_user_id" validate:"required"`
	DueDate        *time.Time `json:"due_date"`
	Keterangan     string     `json:"keterangan"`
}

type UpdateIssueRequest struct {
	IssuePICUserID string      `json:"issue_pic_user_id"`
	DueDate        *time.Time  `json:"due_date"`
	IssueStatus    IssueStatus `json:"issue_status" validate:"omitempty,oneof=Open InProgress PendingValidation Closed Verified"`
	Keterangan     string      `json:"keterangan"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type IssueRepository interface {
	FindAll(page, limit int, status, picUserID string) ([]Issue, int64, error)
	FindByID(id string) (*Issue, error)
	FindByResultID(resultID string) (*Issue, error)
	Create(i *Issue) error
	Update(i *Issue) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type IssueUseCase interface {
	GetAll(page, limit int, status, picUserID string) ([]Issue, int64, error)
	GetByID(id string) (*Issue, error)
	Create(actorID string, req *CreateIssueRequest) (*Issue, error)
	Update(id string, actorID string, req *UpdateIssueRequest) (*Issue, error)
	ExtendDueDate(id string, actorID string, newDueDate time.Time) (*Issue, error)
	Delete(id string, actorID string) error
}

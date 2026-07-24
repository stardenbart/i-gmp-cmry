package issue

import "time"

// IssueDelegate represents a user who is delegated to follow up on an issue
// even if they are not the primary PIC.
type IssueDelegate struct {
	IssueID        string    `gorm:"column:IssueID;primaryKey" json:"issue_id"`
	DelegateUserID string    `gorm:"column:DelegateUserID;primaryKey" json:"delegate_user_id"`
	DelegatedAt    time.Time `gorm:"column:DelegatedAt;autoCreateTime" json:"delegated_at"`
	DelegatedBy    string    `gorm:"column:DelegatedBy;not null" json:"delegated_by"` // User ID of Admin/Auditor who delegated this
}

func (IssueDelegate) TableName() string { return "Issue_Delegate" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

type AddIssueDelegateRequest struct {
	DelegateUserID string `json:"delegate_user_id" validate:"required"`
}

// ─── Repository Interface ──────────────────────────────────────────────────
// (Appended to IssueRepository or separate, but we define separate interface here for clarity)
type IssueDelegateRepository interface {
	AddDelegate(d *IssueDelegate) error
	RemoveDelegate(issueID string, delegateUserID string) error
	GetDelegatesByIssue(issueID string) ([]IssueDelegate, error)
	IsDelegate(issueID string, userID string) (bool, error)
}

// ─── UseCase Interface ─────────────────────────────────────────────────────
type IssueDelegateUseCase interface {
	AddDelegate(issueID string, actorID string, req *AddIssueDelegateRequest) error
	RemoveDelegate(issueID string, delegateUserID string, actorID string) error
	GetDelegatesByIssue(issueID string) ([]IssueDelegate, error)
}

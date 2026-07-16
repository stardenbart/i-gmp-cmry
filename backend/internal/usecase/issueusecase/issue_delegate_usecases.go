package issueusecase

import (
	"errors"
	"time"

	"github.com/monitoring-system/backend/internal/domain/issue"
)

type issueDelegateUseCase struct {
	repo issue.IssueDelegateRepository
}

func NewIssueDelegateUseCase(repo issue.IssueDelegateRepository) issue.IssueDelegateUseCase {
	return &issueDelegateUseCase{repo: repo}
}

func (uc *issueDelegateUseCase) AddDelegate(issueID string, actorID string, req *issue.AddIssueDelegateRequest) error {
	// Optional: Check if actorID is Admin/Auditor (usually done in middleware)
	// Add the delegate
	d := &issue.IssueDelegate{
		IssueID:        issueID,
		DelegateUserID: req.DelegateUserID,
		DelegatedAt:    time.Now(),
		DelegatedBy:    actorID,
	}
	// Avoid duplicates
	isDel, _ := uc.repo.IsDelegate(issueID, req.DelegateUserID)
	if isDel {
		return errors.New("user is already a delegate for this issue")
	}

	return uc.repo.AddDelegate(d)
}

func (uc *issueDelegateUseCase) RemoveDelegate(issueID string, delegateUserID string, actorID string) error {
	return uc.repo.RemoveDelegate(issueID, delegateUserID)
}

func (uc *issueDelegateUseCase) GetDelegatesByIssue(issueID string) ([]issue.IssueDelegate, error) {
	return uc.repo.GetDelegatesByIssue(issueID)
}

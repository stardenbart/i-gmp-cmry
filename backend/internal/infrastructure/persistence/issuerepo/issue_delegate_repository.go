package issuerepo

import (
	"errors"

	"github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
)

type issueDelegateRepo struct {
	db *gorm.DB
}

func NewIssueDelegateRepository(db *gorm.DB) issue.IssueDelegateRepository {
	// Auto migrate the new table
	_ = db.AutoMigrate(&issue.IssueDelegate{})
	return &issueDelegateRepo{db: db}
}

func (r *issueDelegateRepo) AddDelegate(d *issue.IssueDelegate) error {
	return r.db.Create(d).Error
}

func (r *issueDelegateRepo) RemoveDelegate(issueID string, delegateUserID string) error {
	return r.db.Where("IssueID = ? AND DelegateUserID = ?", issueID, delegateUserID).Delete(&issue.IssueDelegate{}).Error
}

func (r *issueDelegateRepo) GetDelegatesByIssue(issueID string) ([]issue.IssueDelegate, error) {
	var delegates []issue.IssueDelegate
	err := r.db.Where("IssueID = ?", issueID).Find(&delegates).Error
	return delegates, err
}

func (r *issueDelegateRepo) IsDelegate(issueID string, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&issue.IssueDelegate{}).Where("IssueID = ? AND DelegateUserID = ?", issueID, userID).Count(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

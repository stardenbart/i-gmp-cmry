package issuerepo

import (
	"github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
)

// ── Issue Repository ──────────────────────────────────────────────────────

type issueRepository struct{ db *gorm.DB }

func NewIssueRepository(db *gorm.DB) issue.IssueRepository {
	return &issueRepository{db: db}
}

func (r *issueRepository) FindAll(page, limit int, status, picUserID string) ([]issue.Issue, int64, error) {
	var items []issue.Issue
	var total int64
	q := r.db.Model(&issue.Issue{})
	if status != "" { q = q.Where("IssueStatus = ?", status) }
	if picUserID != "" { 
		q = q.Where("IssuePICUserID = ? OR IssueID IN (SELECT IssueID FROM Issue_Delegate WHERE DelegateUserID = ?)", picUserID, picUserID) 
	}
	q.Count(&total)
	err := q.Preload("Photos").Order("IssueCreatedAt DESC").Offset((page-1)*limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *issueRepository) FindByID(id string) (*issue.Issue, error) {
	var item issue.Issue
	err := r.db.Preload("Photos").Where("IssueID = ?", id).First(&item).Error
	return &item, err
}

func (r *issueRepository) FindByResultID(resultID string) (*issue.Issue, error) {
	var item issue.Issue
	err := r.db.Where("ResultID = ?", resultID).First(&item).Error
	return &item, err
}

func (r *issueRepository) Create(i *issue.Issue) error   { return r.db.Create(i).Error }
func (r *issueRepository) Update(i *issue.Issue) error   { return r.db.Save(i).Error }
func (r *issueRepository) Delete(id string) error {
	return r.db.Where("IssueID = ?", id).Delete(&issue.Issue{}).Error
}

// ── Issue Photo Repository ─────────────────────────────────────────────────

type issuePhotoRepository struct{ db *gorm.DB }

func NewIssuePhotoRepository(db *gorm.DB) issue.IssuePhotoRepository {
	return &issuePhotoRepository{db: db}
}

func (r *issuePhotoRepository) FindByIssueID(issueID string) ([]issue.IssuePhoto, error) {
	var items []issue.IssuePhoto
	err := r.db.Where("IssueID = ?", issueID).Order("PhotoCreatedAt ASC").Find(&items).Error
	return items, err
}

func (r *issuePhotoRepository) FindByID(id string) (*issue.IssuePhoto, error) {
	var item issue.IssuePhoto
	err := r.db.Where("IssuePhotoID = ?", id).First(&item).Error
	return &item, err
}

func (r *issuePhotoRepository) Create(p *issue.IssuePhoto) error { return r.db.Create(p).Error }
func (r *issuePhotoRepository) Delete(id string) error {
	return r.db.Where("IssuePhotoID = ?", id).Delete(&issue.IssuePhoto{}).Error
}

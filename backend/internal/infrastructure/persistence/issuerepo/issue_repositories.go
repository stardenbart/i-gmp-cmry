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

func (r *issueRepository) FindAll(page, limit int, status, picUserID string, needsWOWR *bool) ([]issue.Issue, int64, error) {
	var items []issue.Issue
	var total int64

	// Base query (no joins to avoid row inflation)
	q := r.db.Model(&issue.Issue{})

	if status != "" {
		q = q.Where(`"Issue"."IssueStatus" = ?`, status)
	}
	if picUserID != "" {
		q = q.Where(
			`"Issue"."IssuePICUserID" = ? 
			OR "Issue"."IssueID" IN (SELECT "IssueID" FROM "Issue_Delegate" WHERE "DelegateUserID" = ?)
			OR "Issue"."IssueID" IN (
				SELECT i."IssueID" FROM "Issue" i
				JOIN "Inspection_Result" ir ON i."ResultID" = ir."ResultID"
				JOIN "Inspection_Header" ih ON ir."InspectionID" = ih."InspectionID"
				JOIN "PIC_Mapping" pm ON ih."KawasanID" = pm."KawasanID"
				WHERE pm."UserID" = ?
			)`,
			picUserID, picUserID, picUserID,
		)
	}
	if needsWOWR != nil {
		q = q.Where(`"Issue"."NeedsWOWR" = ?`, *needsWOWR)
	}

	// Count distinct before pagination
	countQ := q.Session(&gorm.Session{})
	countQ.Select(`COUNT(DISTINCT "IssueID")`).Count(&total)

	// Fetch using subqueries for names — no JOINs that could inflate rows
	err := q.
		Select(`"Issue".*,
			(SELECT am."AreaName" FROM "Inspection_Result" ir2
			  JOIN "Inspection_Header" ih2 ON ih2."InspectionID" = ir2."InspectionID"
			  JOIN "Area_Master" am ON am."AreaID" = ih2."AreaID"
			  WHERE ir2."ResultID" = "Issue"."ResultID" LIMIT 1) AS "AreaName",
			(SELECT km."KawasanName" FROM "Inspection_Result" ir2
			  JOIN "Inspection_Header" ih2 ON ih2."InspectionID" = ir2."InspectionID"
			  JOIN "Kawasan_Master" km ON km."KawasanID" = ih2."KawasanID"
			  WHERE ir2."ResultID" = "Issue"."ResultID" LIMIT 1) AS "KawasanName",
			(SELECT dkm."DetailKawasanName" FROM "Inspection_Result" ir2
			  JOIN "Inspection_Header" ih2 ON ih2."InspectionID" = ir2."InspectionID"
			  JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih2."DetailKawasanID"
			  WHERE ir2."ResultID" = "Issue"."ResultID" LIMIT 1) AS "DetailKawasanName",
			(SELECT u."FullName" FROM "Users" u WHERE u."UserID" = "Issue"."IssuePICUserID" LIMIT 1) AS "PICName"`).
		Preload("Photos").
		Order(`"Issue"."IssueCreatedAt" DESC`).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&items).Error
	return items, total, err
}

func (r *issueRepository) FindByID(id string) (*issue.Issue, error) {
	var item issue.Issue
	err := r.db.Model(&issue.Issue{}).
		Select(`"Issue".*, 
			am."AreaName" AS "AreaName", 
			km."KawasanName" AS "KawasanName", 
			dkm."DetailKawasanName" AS "DetailKawasanName", 
			u."FullName" AS "PICName"`).
		Joins(`LEFT JOIN "Inspection_Result" ir ON ir."ResultID" = "Issue"."ResultID"`).
		Joins(`LEFT JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Joins(`LEFT JOIN "Area_Master" am ON am."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = "Issue"."IssuePICUserID"`).
		Preload("Photos").
		Where(`"Issue"."IssueID" = ?`, id).
		First(&item).Error
	return &item, err
}

func (r *issueRepository) FindByResultID(resultID string) (*issue.Issue, error) {
	var item issue.Issue
	err := r.db.Where("\"ResultID\" = ?", resultID).First(&item).Error
	return &item, err
}

func (r *issueRepository) Create(i *issue.Issue) error { return r.db.Create(i).Error }
func (r *issueRepository) Update(i *issue.Issue) error { return r.db.Save(i).Error }
func (r *issueRepository) Delete(id string) error {
	return r.db.Where("\"IssueID\" = ?", id).Delete(&issue.Issue{}).Error
}

// ── Issue Photo Repository ─────────────────────────────────────────────────

type issuePhotoRepository struct{ db *gorm.DB }

func NewIssuePhotoRepository(db *gorm.DB) issue.IssuePhotoRepository {
	return &issuePhotoRepository{db: db}
}

func (r *issuePhotoRepository) FindByIssueID(issueID string) ([]issue.IssuePhoto, error) {
	var items []issue.IssuePhoto
	err := r.db.Where("\"IssueID\" = ?", issueID).Order(`"PhotoCreatedAt" ASC`).Find(&items).Error
	return items, err
}

func (r *issuePhotoRepository) FindByID(id string) (*issue.IssuePhoto, error) {
	var item issue.IssuePhoto
	err := r.db.Where("\"IssuePhotoID\" = ?", id).First(&item).Error
	return &item, err
}

func (r *issuePhotoRepository) Create(p *issue.IssuePhoto) error { return r.db.Create(p).Error }
func (r *issuePhotoRepository) Update(p *issue.IssuePhoto) error { return r.db.Save(p).Error }
func (r *issuePhotoRepository) Delete(id string) error {
	return r.db.Where("\"IssuePhotoID\" = ?", id).Delete(&issue.IssuePhoto{}).Error
}

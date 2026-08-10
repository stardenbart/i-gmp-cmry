package issuerepo

import (
	"github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ── Issue Repository ──────────────────────────────────────────────────────

type issueRepository struct{ db *gorm.DB }

func NewIssueRepository(db *gorm.DB) issue.IssueRepository {
	return &issueRepository{db: db}
}

func (r *issueRepository) FindAll(page, limit int, plantID, status, picUserID string, needsWOWR *bool) ([]issue.Issue, int64, error) {
	var items []issue.Issue
	var total int64

	// Base query (no joins to avoid row inflation)
	q := r.db.Model(&issue.Issue{})

	if plantID != "" {
		q = q.Where(
			`"IssueID" IN (
				SELECT i."IssueID" FROM "Issue" i
				JOIN "Inspection_Result" ir ON i."ResultID" = ir."ResultID"
				JOIN "Inspection_Header" ih ON ir."InspectionID" = ih."InspectionID"
				JOIN "Area_Master" am ON ih."AreaID" = am."AreaID"
				WHERE am."PlantID" = ?
			)`,
			plantID,
		)
	}

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
			(SELECT u."FullName" FROM "Users" u WHERE u."UserID" = "Issue"."IssuePICUserID" LIMIT 1) AS "PICName",
			(SELECT asp."AspekName" FROM "Inspection_Result" ir2
			  JOIN "Uraian_Master" um ON um."UraianID" = ir2."UraianID"
			  JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"
			  JOIN "Aspek_Master" asp ON asp."AspekID" = dm."AspekID"
			  WHERE ir2."ResultID" = "Issue"."ResultID" LIMIT 1) AS "AspekName",
			(SELECT dm."DetailName" FROM "Inspection_Result" ir2
			  JOIN "Uraian_Master" um ON um."UraianID" = ir2."UraianID"
			  JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"
			  WHERE ir2."ResultID" = "Issue"."ResultID" LIMIT 1) AS "DetailAspekName",
			(SELECT um."UraianText" FROM "Inspection_Result" ir2
			  JOIN "Uraian_Master" um ON um."UraianID" = ir2."UraianID"
			  WHERE ir2."ResultID" = "Issue"."ResultID" LIMIT 1) AS "UraianText",
			(SELECT hm."HabitName" FROM "Issue_HEI" hei
			  JOIN "Habit_Master" hm ON hm."HabitID" = hei."HabitID"
			  WHERE hei."IssueID" = "Issue"."IssueID" LIMIT 1) AS "HabitName",
			(SELECT em."EquipmentName" FROM "Issue_HEI" hei
			  JOIN "Equipment_Master" em ON em."EquipmentID" = hei."EquipmentID"
			  WHERE hei."IssueID" = "Issue"."IssueID" LIMIT 1) AS "EquipmentName",
			(SELECT im."InfrastructureName" FROM "Issue_HEI" hei
			  JOIN "Infrastructure_Master" im ON im."InfrastructureID" = hei."InfrastructureID"
			  WHERE hei."IssueID" = "Issue"."IssueID" LIMIT 1) AS "InfrastructureName"`).
		Preload("Photos").
		Preload("HEI.Habit").
		Preload("HEI.Equipment").
		Preload("HEI.Infrastructure").
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
			u."FullName" AS "PICName",
			asp."AspekName" AS "AspekName",
			dm."DetailName" AS "DetailAspekName",
			um."UraianText" AS "UraianText",
			hm."HabitName" AS "HabitName",
			em."EquipmentName" AS "EquipmentName",
			im."InfrastructureName" AS "InfrastructureName"`).
		Joins(`LEFT JOIN "Inspection_Result" ir ON ir."ResultID" = "Issue"."ResultID"`).
		Joins(`LEFT JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Joins(`LEFT JOIN "Area_Master" am ON am."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = "Issue"."IssuePICUserID"`).
		Joins(`LEFT JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"`).
		Joins(`LEFT JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"`).
		Joins(`LEFT JOIN "Aspek_Master" asp ON asp."AspekID" = dm."AspekID"`).
		Joins(`LEFT JOIN "Issue_HEI" hei ON hei."IssueID" = "Issue"."IssueID"`).
		Joins(`LEFT JOIN "Habit_Master" hm ON hm."HabitID" = hei."HabitID"`).
		Joins(`LEFT JOIN "Equipment_Master" em ON em."EquipmentID" = hei."EquipmentID"`).
		Joins(`LEFT JOIN "Infrastructure_Master" im ON im."InfrastructureID" = hei."InfrastructureID"`).
		Preload("Photos").
		Preload("HEI.Habit").
		Preload("HEI.Equipment").
		Preload("HEI.Infrastructure").
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

// ── Issue HEI Repository ───────────────────────────────────────────────────

type issueHEIRepository struct{ db *gorm.DB }

func NewIssueHEIRepository(db *gorm.DB) issue.IssueHEIRepository {
	_ = db.AutoMigrate(&issue.IssueHEI{})
	return &issueHEIRepository{db: db}
}

func (r *issueHEIRepository) UpsertByIssueID(issueID string, hei *issue.IssueHEI) error {
	hei.IssueID = issueID
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "IssueID"}},
		DoUpdates: clause.AssignmentColumns([]string{"HabitID", "EquipmentID", "InfrastructureID", "UpdatedAt"}),
	}).Create(hei).Error
}

func (r *issueHEIRepository) FindByIssueID(issueID string) (*issue.IssueHEI, error) {
	var item issue.IssueHEI
	err := r.db.Where(`"IssueID" = ?`, issueID).
		Preload("Habit").
		Preload("Equipment").
		Preload("Infrastructure").
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *issueHEIRepository) DeleteByIssueID(issueID string) error {
	return r.db.Where(`"IssueID" = ?`, issueID).Delete(&issue.IssueHEI{}).Error
}

// ── Issue Photo Repository ─────────────────────────────────────────────────

type issuePhotoRepository struct{ db *gorm.DB }

func NewIssuePhotoRepository(db *gorm.DB) issue.IssuePhotoRepository {
	_ = db.AutoMigrate(&issue.IssuePhoto{})
	return &issuePhotoRepository{db: db}
}

func (r *issuePhotoRepository) FindByIssueID(issueID string) ([]issue.IssuePhoto, error) {
	var items []issue.IssuePhoto
	err := r.db.Model(&issue.IssuePhoto{}).
		Select(`"Issue_Photo".*, 
			COALESCE(u."FullName", "Issue_Photo"."PICUserID") AS "UploaderName",
			hei."HEIName" AS "HEIName",
			hei."CategoryName" AS "HEICategory"`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = "Issue_Photo"."PICUserID"`).
		Joins(`LEFT JOIN "HEI_Master" hei ON hei."HEIID" = "Issue_Photo"."HEIID"`).
		Where(`"Issue_Photo"."IssueID" = ?`, issueID).
		Order(`"Issue_Photo"."PhotoCreatedAt" ASC`).
		Find(&items).Error
	return items, err
}

func (r *issuePhotoRepository) FindByID(id string) (*issue.IssuePhoto, error) {
	var item issue.IssuePhoto
	err := r.db.Model(&issue.IssuePhoto{}).
		Select(`"Issue_Photo".*, 
			COALESCE(u."FullName", "Issue_Photo"."PICUserID") AS "UploaderName",
			hei."HEIName" AS "HEIName",
			hei."CategoryName" AS "HEICategory"`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = "Issue_Photo"."PICUserID"`).
		Joins(`LEFT JOIN "HEI_Master" hei ON hei."HEIID" = "Issue_Photo"."HEIID"`).
		Where(`"Issue_Photo"."IssuePhotoID" = ?`, id).
		First(&item).Error
	return &item, err
}

func (r *issuePhotoRepository) Create(p *issue.IssuePhoto) error {
	if (p.HEIID == nil || *p.HEIID == "") && p.HEICategory != "" {
		var heiItem struct{ HEIID string }
		if err := r.db.Table("HEI_Master").Select(`"HEIID"`).Where(`LOWER("CategoryName") = LOWER(?) AND "Status" = 'Active'`, p.HEICategory).Order(`"CreatedAt" ASC`).First(&heiItem).Error; err == nil && heiItem.HEIID != "" {
			p.HEIID = &heiItem.HEIID
		}
	}
	return r.db.Create(p).Error
}
func (r *issuePhotoRepository) Update(p *issue.IssuePhoto) error { return r.db.Save(p).Error }
func (r *issuePhotoRepository) Delete(id string) error {
	return r.db.Where("\"IssuePhotoID\" = ?", id).Delete(&issue.IssuePhoto{}).Error
}

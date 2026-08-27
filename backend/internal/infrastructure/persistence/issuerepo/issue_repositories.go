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
		if *needsWOWR {
			q = q.Where(`
				"Issue"."NeedsWOWR" = TRUE
				OR COALESCE("Issue"."WO_ID", '') <> ''
				OR COALESCE("Issue"."WR_ID", '') <> ''
				OR EXISTS (
					SELECT 1 FROM "Issue_Photo" ip_wowr
					WHERE ip_wowr."IssueID" = "Issue"."IssueID"
					  AND (
						ip_wowr."NeedsWOWR" = TRUE
						OR COALESCE(ip_wowr."WO_ID", '') <> ''
						OR COALESCE(ip_wowr."WR_ID", '') <> ''
					  )
				)
			`)
		} else {
			q = q.Where(`
				"Issue"."NeedsWOWR" = FALSE
				AND COALESCE("Issue"."WO_ID", '') = ''
				AND COALESCE("Issue"."WR_ID", '') = ''
				AND NOT EXISTS (
					SELECT 1 FROM "Issue_Photo" ip_wowr
					WHERE ip_wowr."IssueID" = "Issue"."IssueID"
					  AND (
						ip_wowr."NeedsWOWR" = TRUE
						OR COALESCE(ip_wowr."WO_ID", '') <> ''
						OR COALESCE(ip_wowr."WR_ID", '') <> ''
					  )
				)
			`)
		}
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
			(SELECT COALESCE(NULLIF(hei."CategoryName", ''), ip."HEICategory") 
			  FROM "Issue_Photo" ip 
			  LEFT JOIN "HEI_Master" hei ON hei."HEIID" = ip."HEIID" 
			  WHERE ip."IssueID" = "Issue"."IssueID" 
			  ORDER BY ip."PhotoCreatedAt" ASC LIMIT 1) AS "HEICategory",
			(SELECT hei."HEIName" 
			  FROM "Issue_Photo" ip 
			  JOIN "HEI_Master" hei ON hei."HEIID" = ip."HEIID" 
			  WHERE ip."IssueID" = "Issue"."IssueID" 
			  ORDER BY ip."PhotoCreatedAt" ASC LIMIT 1) AS "HEIName"`).
		Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"PhotoCreatedAt" ASC`)
		}).
		Preload("HEI.HEI").
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
			um."UraianText" AS "UraianText"`).
		Joins(`LEFT JOIN "Inspection_Result" ir ON ir."ResultID" = "Issue"."ResultID"`).
		Joins(`LEFT JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Joins(`LEFT JOIN "Area_Master" am ON am."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = "Issue"."IssuePICUserID"`).
		Joins(`LEFT JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"`).
		Joins(`LEFT JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"`).
		Joins(`LEFT JOIN "Aspek_Master" asp ON asp."AspekID" = dm."AspekID"`).
		Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"PhotoCreatedAt" ASC`)
		}).
		Preload("HEI.HEI").
		Where(`"Issue"."IssueID" = ?`, id).
		Take(&item).Error

	if err == nil && len(item.Photos) > 0 {
		firstPhoto := item.Photos[0]
		item.HEICategory = firstPhoto.HEICategory
		item.HEIName = firstPhoto.HEIName
	}

	return &item, err
}

func (r *issueRepository) FindByResultID(resultID string) (*issue.Issue, error) {
	var item issue.Issue
	err := r.db.Where("\"ResultID\" = ?", resultID).Take(&item).Error
	return &item, err
}

func (r *issueRepository) FindActiveByUraianAndDetailKawasan(uraianID, detailKawasanID string) (*issue.Issue, error) {
	var item issue.Issue
	err := r.db.Model(&issue.Issue{}).
		Joins(`JOIN "Inspection_Result" ir ON ir."ResultID" = "Issue"."ResultID"`).
		Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Where(`ir."UraianID" = ? AND ih."DetailKawasanID" = ? AND "Issue"."IssueStatus" NOT IN (?, ?)`,
			uraianID, detailKawasanID, string(issue.IssueStatusClosed), string(issue.IssueStatusVerified)).
		Order(`"Issue"."IssueCreatedAt" DESC`).
		First(&item).Error
	return &item, err
}

func (r *issueRepository) FindActiveByResultContext(resultID string) (*issue.Issue, error) {
	var detailKawasanID string
	_ = r.db.Table(`"Inspection_Result"`).Select(`ih."DetailKawasanID"`).Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = "Inspection_Result"."InspectionID"`).Where(`"Inspection_Result"."ResultID" = ?`, resultID).Scan(&detailKawasanID).Error
	if detailKawasanID != "" {
		var item issue.Issue
		err := r.db.Model(&issue.Issue{}).Where(`"DetailKawasanID" = ?`, detailKawasanID).Order(`"IssueCreatedAt" ASC`).First(&item).Error
		if err == nil {
			return &item, nil
		}
	}
	var item issue.Issue
	err := r.db.Model(&issue.Issue{}).Select(`"Issue".*`).Joins(`JOIN "Inspection_Result" ir_target ON ir_target."ResultID" = ?`, resultID).Joins(`JOIN "Inspection_Header" ih_target ON ih_target."InspectionID" = ir_target."InspectionID"`).Joins(`JOIN "Inspection_Result" ir_existing ON ir_existing."ResultID" = "Issue"."ResultID"`).Joins(`JOIN "Inspection_Header" ih_existing ON ih_existing."InspectionID" = ir_existing."InspectionID"`).Where(`ih_existing."DetailKawasanID" = ih_target."DetailKawasanID"`).Order(`"Issue"."IssueCreatedAt" ASC`).First(&item).Error
	if err != nil {
		return nil, nil
	}
	return &item, err
}

type dupKawasanGroup struct {
	DetailKawasanID string `gorm:"column:DetailKawasanID"`
}

func (r *issueRepository) ConsolidateDuplicateActiveIssues() error {
	var groups []dupKawasanGroup
	errGroup := r.db.Table(`"Issue" i`).
		Select(`ih."DetailKawasanID" AS "DetailKawasanID"`).
		Joins(`JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"`).
		Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Where(`i."IssueStatus" NOT IN (?, ?)`, string(issue.IssueStatusClosed), string(issue.IssueStatusVerified)).
		Group(`ih."DetailKawasanID"`).
		Having(`COUNT(i."IssueID") > 1`).
		Scan(&groups).Error

	if errGroup != nil || len(groups) == 0 {
		return errGroup
	}

	for _, g := range groups {
		if g.DetailKawasanID == "" {
			continue
		}

		var activeList []issue.Issue
		errFind := r.db.Model(&issue.Issue{}).
			Select(`"Issue".*`).
			Joins(`JOIN "Inspection_Result" ir ON ir."ResultID" = "Issue"."ResultID"`).
			Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
			Where(`ih."DetailKawasanID" = ? AND "Issue"."IssueStatus" NOT IN (?, ?)`,
				g.DetailKawasanID, string(issue.IssueStatusClosed), string(issue.IssueStatusVerified)).
			Order(`"Issue"."IssueCreatedAt" DESC`).
			Find(&activeList).Error

		if errFind != nil || len(activeList) <= 1 {
			continue
		}

		primaryIssue := activeList[0]
		for i := 1; i < len(activeList); i++ {
			dup := activeList[i]
			// Migrate any photos from duplicate issue to primary issue
			_ = r.db.Model(&issue.IssuePhoto{}).
				Where(`"IssueID" = ?`, dup.IssueID).
				Update("IssueID", primaryIssue.IssueID).Error

			// Migrate HEI record if primary does not have one
			_ = r.db.Exec(`UPDATE "Issue_HEI" SET "IssueID" = ? WHERE "IssueID" = ? AND NOT EXISTS (SELECT 1 FROM "Issue_HEI" WHERE "IssueID" = ?)`,
				primaryIssue.IssueID, dup.IssueID, primaryIssue.IssueID).Error

			// Mark duplicate issue as Closed
			dup.IssueStatus = issue.IssueStatusClosed
			_ = r.db.Save(&dup).Error
		}
	}
	return nil
}

func (r *issueRepository) Create(i *issue.Issue) error {
	// Persist the location context together with the issue. This keeps an issue
	// visible and deduplicatable even if its original inspection result is later
	// removed.
	if i.DetailKawasanID == "" && i.ResultID != "" {
		_ = r.db.Table(`"Inspection_Result" ir`).
			Select(`ih."DetailKawasanID"`).
			Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
			Where(`ir."ResultID" = ?`, i.ResultID).
			Scan(&i.DetailKawasanID).Error
	}
	return r.db.Create(i).Error
}
func (r *issueRepository) Update(i *issue.Issue) error { return r.db.Save(i).Error }
func (r *issueRepository) Delete(id string) error {
	return r.db.Where("\"IssueID\" = ?", id).Delete(&issue.Issue{}).Error
}

// ── Issue HEI Repository ───────────────────────────────────────────────────

type issueHEIRepository struct{ db *gorm.DB }

func NewIssueHEIRepository(db *gorm.DB) issue.IssueHEIRepository {
	return &issueHEIRepository{db: db}
}

func (r *issueHEIRepository) UpsertByIssueID(issueID string, hei *issue.IssueHEI) error {
	hei.IssueID = issueID
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "IssueID"}},
		DoUpdates: clause.AssignmentColumns([]string{"HEIID", "UpdatedAt"}),
	}).Create(hei).Error
}

func (r *issueHEIRepository) FindByIssueID(issueID string) (*issue.IssueHEI, error) {
	var item issue.IssueHEI
	err := r.db.Where(`"IssueID" = ?`, issueID).
		Preload("HEI").
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
	return &issuePhotoRepository{db: db}
}

func (r *issuePhotoRepository) FindByIssueID(issueID string) ([]issue.IssuePhoto, error) {
	var items []issue.IssuePhoto
	err := r.db.Model(&issue.IssuePhoto{}).
		Select(`"Issue_Photo".*, 
			COALESCE(u."FullName", "Issue_Photo"."PICUserID") AS "UploaderName",
			hei."HEIName" AS "HEIName",
			COALESCE(NULLIF(hei."CategoryName", ''), "Issue_Photo"."HEICategory") AS "HEICategory"`).
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
			COALESCE(NULLIF(hei."CategoryName", ''), "Issue_Photo"."HEICategory") AS "HEICategory"`).
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
func (r *issuePhotoRepository) Update(p *issue.IssuePhoto) error {
	if (p.HEIID == nil || *p.HEIID == "") && p.HEICategory != "" {
		var heiItem struct{ HEIID string }
		if err := r.db.Table("HEI_Master").Select(`"HEIID"`).Where(`LOWER("CategoryName") = LOWER(?) AND "Status" = 'Active'`, p.HEICategory).Order(`"CreatedAt" ASC`).First(&heiItem).Error; err == nil && heiItem.HEIID != "" {
			p.HEIID = &heiItem.HEIID
		}
	}
	return r.db.Save(p).Error
}
func (r *issuePhotoRepository) Delete(id string) error {
	return r.db.Where("\"IssuePhotoID\" = ?", id).Delete(&issue.IssuePhoto{}).Error
}

func (r *issueRepository) FindByDetailKawasanID(detailKawasanID string) (*issue.Issue, error) {
	var item issue.Issue
	err := r.db.Model(&issue.Issue{}).Where(`"DetailKawasanID" = ?`, detailKawasanID).Order(`"IssueCreatedAt" ASC`).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

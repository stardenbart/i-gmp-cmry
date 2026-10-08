package inspectionrepo

import (
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"gorm.io/gorm"
)

type kawasanReportRepository struct {
	db *gorm.DB
}

func NewKawasanReportRepository(db *gorm.DB) inspection.KawasanReportRepository {
	return &kawasanReportRepository{db: db}
}

// kawasanReportSQL takes, per Detail Kawasan, the latest Completed/Approved
// inspection created in [start, end) — the same rule CountCompletedInPeriod
// uses to decide the Kawasan is fully inspected — and returns its results with
// any finding and its latest follow-up.
const kawasanReportSQL = `
WITH latest AS (
	SELECT DISTINCT ON (ih."DetailKawasanID")
		ih."InspectionID", ih."DetailKawasanID", ih."AreaID", ih."KawasanID"
	FROM "Inspection_Header" ih
	WHERE ih."KawasanID" = ?
	  AND ih."InspectionHeaderStatus" IN ('Completed', 'Approved')
	  AND ih."InspectionHeaderCreatedAt" >= ? AND ih."InspectionHeaderCreatedAt" < ?
	ORDER BY ih."DetailKawasanID", ih."InspectionHeaderCreatedAt" DESC
)
SELECT
	COALESCE(am."AreaName", l."AreaID") AS area_name,
	COALESCE(km."KawasanName", l."KawasanID") AS kawasan_name,
	l."DetailKawasanID" AS detail_kawasan_id,
	COALESCE(dkm."DetailKawasanName", l."DetailKawasanID") AS detail_kawasan_name,
	COALESCE(ir."Nilai", 0) AS nilai,
	COALESCE(um."StandardScore", 0) AS standard_score,
	COALESCE(um."UraianText", '') AS uraian_text,
	i."IssueID" AS issue_id,
	COALESCE(i."Keterangan", '') AS keterangan,
	COALESCE(i."IssueStatus", '') AS issue_status,
	i."DueDate" AS due_date,
	fu.follow_up_by,
	fu.follow_up_date
FROM latest l
-- LEFT JOIN keeps a Detail Kawasan whose inspection has no checklist rows.
LEFT JOIN "Inspection_Result" ir ON ir."InspectionID" = l."InspectionID"
LEFT JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"
LEFT JOIN "Area_Master" am ON am."AreaID" = l."AreaID"
LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = l."KawasanID"
LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = l."DetailKawasanID"
LEFT JOIN "Issue" i ON i."ResultID" = ir."ResultID"
LEFT JOIN LATERAL (
	SELECT COALESCE(fu_user."FullName", p."PICUserID") AS follow_up_by,
	       COALESCE(p."FollowUpDate", p."PhotoCreatedAt") AS follow_up_date
	FROM "Issue_Photo" p
	LEFT JOIN "Users" fu_user ON fu_user."UserID" = p."PICUserID"
	WHERE p."IssueID" = i."IssueID" AND p."PhotoType" = 'FollowUp'
	ORDER BY COALESCE(p."FollowUpDate", p."PhotoCreatedAt") DESC
	LIMIT 1
) fu ON TRUE
ORDER BY detail_kawasan_name, ir."ResultID"`

func (r *kawasanReportRepository) FindKawasanReportRows(kawasanID string, periodStart, periodEnd time.Time) ([]inspection.KawasanReportRow, error) {
	var rows []inspection.KawasanReportRow
	err := r.db.Raw(kawasanReportSQL, kawasanID, periodStart, periodEnd).Scan(&rows).Error
	return rows, err
}

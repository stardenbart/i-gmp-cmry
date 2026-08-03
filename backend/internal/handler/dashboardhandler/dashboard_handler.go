package dashboardhandler

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/exporter"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/response"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db        *gorm.DB
	log       *logger.Logger
	cryptoSvc *crypto.Service
}

func NewDashboardHandler(db *gorm.DB, log *logger.Logger, cryptoSvc *crypto.Service) *DashboardHandler {
	return &DashboardHandler{
		db:        db,
		log:       log,
		cryptoSvc: cryptoSvc,
	}
}

func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	var totalInspectionsRunning int64
	var totalOpenIssues int64
	var inspectionsCompleted int64
	var inspectionsRunning int64
	var picFollowupCompleted int64
	var picFollowupOverdue int64
	var totalOK, totalCheck int64

	kawasanID := c.Query("kawasan_id")
	areaID := c.Query("area_id")

	var filterAreas []string
	if kawasanID != "" {
		h.db.Table("\"Area_Master\"").Select("\"AreaID\"").Where("\"KawasanID\" = ?", kawasanID).Scan(&filterAreas)
		if len(filterAreas) == 0 {
			filterAreas = []string{"RESTRICTED_NONE"}
		}
	} else if areaID != "" {
		filterAreas = []string{areaID}
	}

	allowedAreas := h.getAllowedAreas(c)
	if allowedAreas != nil {
		if filterAreas != nil {
			var intersected []string
			allowedMap := make(map[string]bool)
			for _, a := range allowedAreas {
				allowedMap[a] = true
			}
			for _, f := range filterAreas {
				if allowedMap[f] {
					intersected = append(intersected, f)
				}
			}
			if len(intersected) == 0 {
				intersected = []string{"RESTRICTED_NONE"}
			}
			allowedAreas = intersected
		}
	} else if filterAreas != nil {
		allowedAreas = filterAreas
	}

	qRunning := h.db.Model(&inspection.InspectionHeader{}).Where("\"InspectionHeaderStatus\" = ?", "Ongoing")
	if allowedAreas != nil { qRunning = qRunning.Where("\"AreaID\" IN ?", allowedAreas) }
	qRunning.Count(&totalInspectionsRunning)

	qCompleted := h.db.Model(&inspection.InspectionHeader{}).Where("\"InspectionHeaderStatus\" = ?", "Completed")
	if allowedAreas != nil { qCompleted = qCompleted.Where("\"AreaID\" IN ?", allowedAreas) }
	qCompleted.Count(&inspectionsCompleted)

	// Issue queries need join with Inspection_Header
	qIssueBase := h.db.Table("\"Issue\" i").Joins("JOIN \"Inspection_Result\" ir ON ir.\"ResultID\" = i.\"ResultID\"").Joins("JOIN \"Inspection_Header\" ih ON ih.\"InspectionID\" = ir.\"InspectionID\"")
	qIssueOpen := qIssueBase.Session(&gorm.Session{}).Where("i.\"IssueStatus\" != ?", "Closed")
	if allowedAreas != nil { qIssueOpen = qIssueOpen.Where("ih.\"AreaID\" IN ?", allowedAreas) }
	qIssueOpen.Count(&totalOpenIssues)

	qIssueClosed := qIssueBase.Session(&gorm.Session{}).Where("i.\"IssueStatus\" = ?", "Closed")
	if allowedAreas != nil { qIssueClosed = qIssueClosed.Where("ih.\"AreaID\" IN ?", allowedAreas) }
	qIssueClosed.Count(&picFollowupCompleted)

	qIssueOverdue := qIssueBase.Session(&gorm.Session{}).Where("i.\"IssueStatus\" != ? AND i.\"DueDate\" < NOW()", "Closed")
	if allowedAreas != nil { qIssueOverdue = qIssueOverdue.Where("ih.\"AreaID\" IN ?", allowedAreas) }
	qIssueOverdue.Count(&picFollowupOverdue)

	qResultBase := h.db.Table("\"Inspection_Result\" ir").Joins("JOIN \"Inspection_Header\" ih ON ih.\"InspectionID\" = ir.\"InspectionID\"")
	qCheck := qResultBase.Session(&gorm.Session{})
	if allowedAreas != nil { qCheck = qCheck.Where("ih.\"AreaID\" IN ?", allowedAreas) }
	qCheck.Count(&totalCheck)

	qOK := qResultBase.Session(&gorm.Session{}).Where("ir.\"Checking\" = ?", "OK")
	if allowedAreas != nil { qOK = qOK.Where("ih.\"AreaID\" IN ?", allowedAreas) }
	qOK.Count(&totalOK)

	inspectionsRunning = totalInspectionsRunning

	complianceRate := 0.0
	if totalCheck > 0 {
		complianceRate = float64(totalOK) / float64(totalCheck) * 100
	}

	// Calculate Auditee Status (Real data from DB)
	type AuditeeStatusRow struct {
		AreaName   string `gorm:"column:name"`
		TotalCheck int64  `gorm:"column:total_check"`
		TotalOK    int64  `gorm:"column:total_ok"`
		OpenIssues int64  `gorm:"column:open_issues"`
	}
	var auditeeRows []AuditeeStatusRow
	queryStr := `
		SELECT 
			am."AreaName" as name,
			COUNT(DISTINCT ir."ResultID") as total_check,
			COUNT(DISTINCT CASE WHEN ir."Checking" = 'OK' THEN ir."ResultID" END) as total_ok,
			COUNT(DISTINCT CASE WHEN i."IssueStatus" != 'Closed' AND i."IssueStatus" != 'Verified' AND i."IssueStatus" IS NOT NULL THEN i."IssueID" END) as open_issues
		FROM "Area_Master" am
		LEFT JOIN "Inspection_Header" ih ON ih."AreaID" = am."AreaID"
		LEFT JOIN "Inspection_Result" ir ON ir."InspectionID" = ih."InspectionID"
		LEFT JOIN "Issue" i ON i."ResultID" = ir."ResultID"
	`
	var args []interface{}
	if allowedAreas != nil {
		queryStr += ` WHERE am."AreaID" IN ? `
		args = append(args, allowedAreas)
	}
	queryStr += ` GROUP BY am."AreaID", am."AreaName" ORDER BY open_issues DESC LIMIT 5 `
	h.db.Raw(queryStr, args...).Scan(&auditeeRows)

	var auditeeStatusList []fiber.Map
	for i, r := range auditeeRows {
		compliance := 0.0
		if r.TotalCheck > 0 {
			compliance = float64(r.TotalOK) / float64(r.TotalCheck) * 100
		}
		
		// For aesthetics, alternate between Store and Factory icons for different areas
		icon := "Store"
		typeStr := "Area Operasional"
		if i%2 != 0 {
			icon = "Factory"
			typeStr = "Area Produksi"
		}

		auditeeStatusList = append(auditeeStatusList, fiber.Map{
			"name":        r.AreaName,
			"type":        typeStr,
			"compliance":  int(compliance), // Return as integer percentage to match frontend
			"open_issues": r.OpenIssues,
			"icon":        icon,
		})
	}

	if len(auditeeStatusList) == 0 {
		auditeeStatusList = []fiber.Map{} // return empty array instead of null
	}

	// Real Multi-Period Historical Compliance Trend Query (1m, 3m, 6m, 1y)
	period := c.Query("period", "6m")
	var complianceTrend []fiber.Map
	now := time.Now()

	switch period {
	case "1m", "1M":
		// 1 Month scope: 4-week breakdown
		for i := 3; i >= 0; i-- {
			endOfWeek := now.AddDate(0, 0, -i*7)
			startOfWeek := endOfWeek.AddDate(0, 0, -6)
			weekLabel := fmt.Sprintf("W%d (%s)", 4-i, startOfWeek.Format("02/01"))

			var mTotalCheck, mTotalOK int64
			qWeekBase := h.db.Table(`"Inspection_Result" ir`).
				Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
				Where(`ih."InspectionHeaderCreatedAt" >= ? AND ih."InspectionHeaderCreatedAt" <= ?`, startOfWeek, endOfWeek)
			if allowedAreas != nil {
				qWeekBase = qWeekBase.Where(`ih."AreaID" IN ?`, allowedAreas)
			}
			qWeekBase.Count(&mTotalCheck)

			qWeekOK := h.db.Table(`"Inspection_Result" ir`).
				Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
				Where(`ih."InspectionHeaderCreatedAt" >= ? AND ih."InspectionHeaderCreatedAt" <= ? AND ir."Checking" = 'OK'`, startOfWeek, endOfWeek)
			if allowedAreas != nil {
				qWeekOK = qWeekOK.Where(`ih."AreaID" IN ?`, allowedAreas)
			}
			qWeekOK.Count(&mTotalOK)

			mRate := complianceRate
			if mTotalCheck > 0 {
				mRate = float64(mTotalOK) / float64(mTotalCheck) * 100
			}

			complianceTrend = append(complianceTrend, fiber.Map{
				"month": weekLabel,
				"rate":  float64(int(mRate*10)) / 10.0,
			})
		}

	case "3m", "3M":
		// 3 Months scope
		for i := 2; i >= 0; i-- {
			targetMonth := now.AddDate(0, -i, 0)
			monthName := targetMonth.Format("Jan")
			yearMonthStr := targetMonth.Format("2006-01")

			var mTotalCheck, mTotalOK int64
			qMonthBase := h.db.Table(`"Inspection_Result" ir`).
				Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
				Where(`TO_CHAR(ih."InspectionHeaderCreatedAt", 'YYYY-MM') = ?`, yearMonthStr)
			if allowedAreas != nil {
				qMonthBase = qMonthBase.Where(`ih."AreaID" IN ?`, allowedAreas)
			}
			qMonthBase.Count(&mTotalCheck)

			qMonthOK := h.db.Table(`"Inspection_Result" ir`).
				Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
				Where(`TO_CHAR(ih."InspectionHeaderCreatedAt", 'YYYY-MM') = ? AND ir."Checking" = 'OK'`, yearMonthStr)
			if allowedAreas != nil {
				qMonthOK = qMonthOK.Where(`ih."AreaID" IN ?`, allowedAreas)
			}
			qMonthOK.Count(&mTotalOK)

			mRate := complianceRate
			if mTotalCheck > 0 {
				mRate = float64(mTotalOK) / float64(mTotalCheck) * 100
			}

			complianceTrend = append(complianceTrend, fiber.Map{
				"month": monthName,
				"rate":  float64(int(mRate*10)) / 10.0,
			})
		}

	case "1y", "1Y", "12m":
		// 1 Year (12 Months) scope
		for i := 11; i >= 0; i-- {
			targetMonth := now.AddDate(0, -i, 0)
			monthName := targetMonth.Format("Jan")
			yearMonthStr := targetMonth.Format("2006-01")

			var mTotalCheck, mTotalOK int64
			qMonthBase := h.db.Table(`"Inspection_Result" ir`).
				Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
				Where(`TO_CHAR(ih."InspectionHeaderCreatedAt", 'YYYY-MM') = ?`, yearMonthStr)
			if allowedAreas != nil {
				qMonthBase = qMonthBase.Where(`ih."AreaID" IN ?`, allowedAreas)
			}
			qMonthBase.Count(&mTotalCheck)

			qMonthOK := h.db.Table(`"Inspection_Result" ir`).
				Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
				Where(`TO_CHAR(ih."InspectionHeaderCreatedAt", 'YYYY-MM') = ? AND ir."Checking" = 'OK'`, yearMonthStr)
			if allowedAreas != nil {
				qMonthOK = qMonthOK.Where(`ih."AreaID" IN ?`, allowedAreas)
			}
			qMonthOK.Count(&mTotalOK)

			mRate := complianceRate
			if mTotalCheck > 0 {
				mRate = float64(mTotalOK) / float64(mTotalCheck) * 100
			}

			complianceTrend = append(complianceTrend, fiber.Map{
				"month": monthName,
				"rate":  float64(int(mRate*10)) / 10.0,
			})
		}

	default: // "6m"
		// 6 Months scope (default)
		for i := 5; i >= 0; i-- {
			targetMonth := now.AddDate(0, -i, 0)
			monthName := targetMonth.Format("Jan")
			yearMonthStr := targetMonth.Format("2006-01")

			var mTotalCheck, mTotalOK int64
			qMonthBase := h.db.Table(`"Inspection_Result" ir`).
				Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
				Where(`TO_CHAR(ih."InspectionHeaderCreatedAt", 'YYYY-MM') = ?`, yearMonthStr)
			if allowedAreas != nil {
				qMonthBase = qMonthBase.Where(`ih."AreaID" IN ?`, allowedAreas)
			}
			qMonthBase.Count(&mTotalCheck)

			qMonthOK := h.db.Table(`"Inspection_Result" ir`).
				Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
				Where(`TO_CHAR(ih."InspectionHeaderCreatedAt", 'YYYY-MM') = ? AND ir."Checking" = 'OK'`, yearMonthStr)
			if allowedAreas != nil {
				qMonthOK = qMonthOK.Where(`ih."AreaID" IN ?`, allowedAreas)
			}
			qMonthOK.Count(&mTotalOK)

			mRate := complianceRate
			if mTotalCheck > 0 {
				mRate = float64(mTotalOK) / float64(mTotalCheck) * 100
			}

			complianceTrend = append(complianceTrend, fiber.Map{
				"month": monthName,
				"rate":  float64(int(mRate*10)) / 10.0,
			})
		}
	}

	// Calculate trend diff vs previous entry
	trendDiff := 0.0
	if len(complianceTrend) >= 2 {
		currRate := complianceTrend[len(complianceTrend)-1]["rate"].(float64)
		prevRate := complianceTrend[len(complianceTrend)-2]["rate"].(float64)
		trendDiff = float64(int((currRate-prevRate)*10)) / 10.0
	}

	data := fiber.Map{
		"total_inspections_running": totalInspectionsRunning,
		"compliance_rate":           complianceRate,
		"compliance_rate_trend":     trendDiff,
		"total_open_issues":         totalOpenIssues,
		"issue_overdue":             picFollowupOverdue,
		"issue_overdue_trend":       0,
		"inspections_completed":     inspectionsCompleted,
		"inspections_running":       inspectionsRunning,
		"pic_followup_completed":    picFollowupCompleted,
		"pic_followup_overdue":      picFollowupOverdue,
		"auditee_status":            auditeeStatusList,
		"compliance_trend":          complianceTrend,
	}

	return c.JSON(fiber.Map{
		"message": "success",
		"data":    data,
	})
}

// ExportStats generates and downloads the global stats as an Excel file using a template.
// GET /api/v1/dashboard/export
func (h *DashboardHandler) GetPreviewExport(c *fiber.Ctx) error {
	areaID := c.Query("area_id")
	kawasanID := c.Query("kawasan_id")
	detailKawasanID := c.Query("detail_kawasan_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	type PreviewExportResponse struct {
		InspectionID    string     `json:"inspection_id"`
		Tanggal         string     `json:"tanggal"`
		AreaID          string     `json:"area_id"`
		Area            string     `json:"area"`
		KawasanID       string     `json:"kawasan_id"`
		Kawasan         string     `json:"kawasan"`
		DetailKawasanID string     `json:"detail_kawasan_id"`
		DetailKawasan   string     `json:"detail_kawasan"`
		PIC             string     `json:"pic"`
		Aspek           string     `json:"aspek"`
		Detail          string     `json:"detail"`
		UraianID        string     `json:"uraian_id"`
		Uraian          string     `json:"uraian"`
		Nilai           int        `json:"nilai"`
		TotalNilai      int        `json:"total_nilai"`
		TotalTemuan     int        `json:"total_temuan"`
		Keterangan      string     `json:"keterangan"`
		IssueID         *string    `json:"issue_id"`
		DueDate         *time.Time `json:"due_date"`
		ImageURL        *string    `json:"image_url"`
		FollowUpDate    *time.Time `json:"follow_up_date"`
	}

	var results []PreviewExportResponse
	query := h.db.Table(`"Inspection_Header" ih`).
		Select(`ih."InspectionID" as inspection_id, 
			ih."InspectionHeaderCreatedAt" as tanggal, 
			ih."AreaID" as area_id, 
			COALESCE(am_area."AreaName", ih."AreaID") as area, 
			ih."KawasanID" as kawasan_id,
			COALESCE(km."KawasanName", ih."KawasanID") as kawasan,
			ih."DetailKawasanID" as detail_kawasan_id,
			COALESCE(dkm."DetailKawasanName", ih."DetailKawasanID") as detail_kawasan,
			COALESCE(u_pic."FullName", u_inspector."FullName", ih."InspectorID") as pic, 
			am."AspekName" as aspek, 
			dm."DetailName" as detail, 
			um."UraianID" as uraian_id, 
			um."UraianText" as uraian, 
			ir."Nilai" as nilai, 
			COALESCE(NULLIF(iss."Keterangan", ''), ir."Keterangan") as keterangan, 
			iss."IssueID" as issue_id, 
			iss."DueDate" as due_date, 
			COALESCE(
				(SELECT p3."ImageUrl" FROM "Issue_Photo" p3 WHERE p3."IssueID" = iss."IssueID" AND p3."PhotoType" IN ('FollowUp', 'WOWR') ORDER BY p3."PhotoCreatedAt" DESC LIMIT 1),
				(SELECT p4."ImageUrl" FROM "Issue_Photo" p4 WHERE p4."IssueID" = iss."IssueID" ORDER BY p4."PhotoCreatedAt" ASC LIMIT 1)
			) as image_url, 
			(SELECT COALESCE(p2."FollowUpDate", p2."PhotoCreatedAt") FROM "Issue_Photo" p2 WHERE p2."IssueID" = iss."IssueID" AND p2."PhotoType" IN ('FollowUp', 'WOWR') ORDER BY p2."PhotoCreatedAt" DESC LIMIT 1) as follow_up_date`).
		Joins(`JOIN "Inspection_Result" ir ON ir."InspectionID" = ih."InspectionID"`).
		Joins(`JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"`).
		Joins(`JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"`).
		Joins(`JOIN "Aspek_Master" am ON am."AspekID" = dm."AspekID"`).
		Joins(`LEFT JOIN "Area_Master" am_area ON am_area."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u_inspector ON u_inspector."UserID" = ih."InspectorID"`).
		Joins(`LEFT JOIN "Issue" iss ON iss."ResultID" = ir."ResultID"`).
		Joins(`LEFT JOIN "Users" u_pic ON u_pic."UserID" = iss."IssuePICUserID"`)

	allowedAreas := h.getAllowedAreas(c)
	if areaID != "" {
		query = query.Where(`ih."AreaID" = ?`, areaID)
	}
	if kawasanID != "" {
		query = query.Where(`ih."KawasanID" = ?`, kawasanID)
	}
	if detailKawasanID != "" {
		query = query.Where(`ih."DetailKawasanID" = ?`, detailKawasanID)
	}
	if startDate != "" {
		query = query.Where(`ih."InspectionHeaderCreatedAt" >= ?`, startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where(`ih."InspectionHeaderCreatedAt" <= ?`, endDate+" 23:59:59")
	}
	if allowedAreas != nil {
		query = query.Where(`ih."AreaID" IN ?`, allowedAreas)
	}
	
	err := query.Order(`ih."InspectionHeaderCreatedAt" DESC`).Scan(&results).Error
	if err != nil {
		h.log.Error("Failed to fetch preview export data", logger.Error(err))
		return response.InternalServerError(c, "Failed to load data", err.Error())
	}

	// Calculate TotalNilai & TotalTemuan per Kawasan and decrypt ImageURL
	kawasanNilai := make(map[string]int)
	kawasanTemuan := make(map[string]int)

	for _, res := range results {
		key := res.KawasanID
		if key == "" {
			key = res.Kawasan
		}
		kawasanNilai[key] += res.Nilai
		if res.IssueID != nil && *res.IssueID != "" {
			kawasanTemuan[key]++
		}
	}

	for i := range results {
		key := results[i].KawasanID
		if key == "" {
			key = results[i].Kawasan
		}
		results[i].TotalNilai = kawasanNilai[key]
		results[i].TotalTemuan = kawasanTemuan[key]

		if h.cryptoSvc != nil && results[i].ImageURL != nil && *results[i].ImageURL != "" {
			decrypted := h.cryptoSvc.DecryptWithFallback(*results[i].ImageURL)
			results[i].ImageURL = &decrypted
		}
		if h.cryptoSvc != nil && results[i].Keterangan != "" {
			results[i].Keterangan = h.cryptoSvc.DecryptWithFallback(results[i].Keterangan)
		}
	}

	return response.OK(c, "success", results)
}

func (h *DashboardHandler) ExportStats(c *fiber.Ctx) error {
	areaID := c.Query("area_id")
	kawasanID := c.Query("kawasan_id")
	detailKawasanID := c.Query("detail_kawasan_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	type PreviewExportResponse struct {
		InspectionID    string     `gorm:"column:inspection_id" json:"inspection_id"`
		Tanggal         time.Time  `gorm:"column:tanggal" json:"tanggal"`
		Area            string     `gorm:"column:area" json:"area"`
		KawasanID       string     `gorm:"column:kawasan_id" json:"kawasan_id"`
		Kawasan         string     `gorm:"column:kawasan" json:"kawasan"`
		DetailKawasanID string     `gorm:"column:detail_kawasan_id" json:"detail_kawasan_id"`
		DetailKawasan   string     `gorm:"column:detail_kawasan" json:"detail_kawasan"`
		PIC             string     `gorm:"column:pic" json:"pic"`
		Aspek           string     `gorm:"column:aspek" json:"aspek"`
		Detail          string     `gorm:"column:detail" json:"detail"`
		UraianID        string     `gorm:"column:uraian_id" json:"uraian_id"`
		Uraian          string     `gorm:"column:uraian" json:"uraian"`
		Nilai           int        `gorm:"column:nilai" json:"nilai"`
		Keterangan      string     `gorm:"column:keterangan" json:"keterangan"`
		IssueID         *string    `gorm:"column:issue_id" json:"issue_id"`
		DueDate         *time.Time `gorm:"column:due_date" json:"due_date"`
		ImageURL        *string    `gorm:"column:image_url" json:"image_url"`
		FollowUpDate    *time.Time `gorm:"column:follow_up_date" json:"follow_up_date"`
	}

	var results []PreviewExportResponse
	query := h.db.Table(`"Inspection_Header" ih`).
		Select(`ih."InspectionID" as inspection_id, 
			ih."InspectionHeaderCreatedAt" as tanggal, 
			COALESCE(am_area."AreaName", ih."AreaID") as area, 
			ih."KawasanID" as kawasan_id,
			COALESCE(km."KawasanName", ih."KawasanID") as kawasan,
			ih."DetailKawasanID" as detail_kawasan_id,
			COALESCE(dkm."DetailKawasanName", ih."DetailKawasanID") as detail_kawasan,
			COALESCE(u_pic."FullName", u_inspector."FullName", ih."InspectorID") as pic, 
			am."AspekName" as aspek, 
			dm."DetailName" as detail, 
			um."UraianID" as uraian_id, 
			um."UraianText" as uraian, 
			ir."Nilai" as nilai, 
			COALESCE(NULLIF(iss."Keterangan", ''), ir."Keterangan") as keterangan, 
			iss."IssueID" as issue_id, 
			iss."DueDate" as due_date, 
			COALESCE(
				(SELECT p3."ImageUrl" FROM "Issue_Photo" p3 WHERE p3."IssueID" = iss."IssueID" AND p3."PhotoType" IN ('FollowUp', 'WOWR') ORDER BY p3."PhotoCreatedAt" DESC LIMIT 1),
				(SELECT p4."ImageUrl" FROM "Issue_Photo" p4 WHERE p4."IssueID" = iss."IssueID" ORDER BY p4."PhotoCreatedAt" ASC LIMIT 1)
			) as image_url, 
			(SELECT COALESCE(p2."FollowUpDate", p2."PhotoCreatedAt") FROM "Issue_Photo" p2 WHERE p2."IssueID" = iss."IssueID" AND p2."PhotoType" IN ('FollowUp', 'WOWR') ORDER BY p2."PhotoCreatedAt" DESC LIMIT 1) as follow_up_date`).
		Joins(`JOIN "Inspection_Result" ir ON ir."InspectionID" = ih."InspectionID"`).
		Joins(`JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"`).
		Joins(`JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"`).
		Joins(`JOIN "Aspek_Master" am ON am."AspekID" = dm."AspekID"`).
		Joins(`LEFT JOIN "Area_Master" am_area ON am_area."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u_inspector ON u_inspector."UserID" = ih."InspectorID"`).
		Joins(`LEFT JOIN "Issue" iss ON iss."ResultID" = ir."ResultID"`).
		Joins(`LEFT JOIN "Users" u_pic ON u_pic."UserID" = iss."IssuePICUserID"`)

	allowedAreas := h.getAllowedAreas(c)
	var areaName string = "Semua Area"
	var picName string = "Semua PIC"
	if areaID != "" {
		query = query.Where(`ih."AreaID" = ?`, areaID)
		h.db.Table(`"Area_Master"`).Select(`"AreaName"`).Where(`"AreaID" = ?`, areaID).Scan(&areaName)
		h.db.Table(`"PIC_Mapping" pm`).Select(`u."FullName"`).Joins(`JOIN "Users" u ON u."UserID" = pm."UserID"`).Where(`pm."AreaID" = ?`, areaID).Limit(1).Scan(&picName)
	}
	if kawasanID != "" {
		query = query.Where(`ih."KawasanID" = ?`, kawasanID)
	}
	if detailKawasanID != "" {
		query = query.Where(`ih."DetailKawasanID" = ?`, detailKawasanID)
	}
	if startDate != "" {
		query = query.Where(`ih."InspectionHeaderCreatedAt" >= ?`, startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where(`ih."InspectionHeaderCreatedAt" <= ?`, endDate+" 23:59:59")
	}
	if allowedAreas != nil {
		query = query.Where(`ih."AreaID" IN ?`, allowedAreas)
	}

	err := query.Order(`ih."InspectionHeaderCreatedAt" DESC`).Scan(&results).Error
	if err != nil {
		h.log.Error("Failed to fetch export stats data", logger.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to fetch export data",
		})
	}

	if picName == "Semua PIC" && len(results) > 0 && results[0].PIC != "" {
		picName = results[0].PIC
	}

	totalNilai := 0
	totalTemuan := 0
	kawasanNilai := make(map[string]int)
	kawasanTemuan := make(map[string]int)

	for i := range results {
		totalNilai += results[i].Nilai
		key := results[i].KawasanID
		if key == "" {
			key = results[i].Kawasan
		}
		kawasanNilai[key] += results[i].Nilai
		if results[i].IssueID != nil && *results[i].IssueID != "" {
			totalTemuan++
			kawasanTemuan[key]++
		}

		if h.cryptoSvc != nil {
			if results[i].ImageURL != nil && *results[i].ImageURL != "" {
				decrypted := h.cryptoSvc.DecryptWithFallback(*results[i].ImageURL)
				results[i].ImageURL = &decrypted
			}
			if results[i].Keterangan != "" {
				results[i].Keterangan = h.cryptoSvc.DecryptWithFallback(results[i].Keterangan)
			}
		}
	}

	// Prepare header DueDate from the first inspection date or leave empty
	headerDueDate := ""
	if len(results) > 0 {
		headerDueDate = results[0].Tanggal.Format("02-Jan-2006")
		// If any result has an issue due date, use the latest one
		for _, r := range results {
			if r.DueDate != nil {
				headerDueDate = r.DueDate.Format("02-Jan-2006")
				break
			}
		}
	}

	// Prepare payload for placeholder exporter
	payload := &exporter.PlaceholderPayload{
		Headers: map[string]interface{}{
			"datetime.now()":      time.Now().Format("02-Jan-06"),
			"AreaName":            areaName,
			"PICName":             picName,
			"DueDate":             headerDueDate,
			"total_semua_nilai":   totalNilai,
			"total_semua_temuan":  totalTemuan,
		},
		Items: make([]map[string]interface{}, 0, len(results)),
	}

	for i, res := range results {
		dueDateStr := ""
		if res.DueDate != nil {
			dueDateStr = res.DueDate.Format("02-Jan-2006")
		}
		followUpStr := ""
		if res.FollowUpDate != nil {
			followUpStr = res.FollowUpDate.Format("02-Jan-2006 15:04")
		}
		imgUrl := ""
		if res.ImageURL != nil && *res.ImageURL != "" {
			imgUrl = *res.ImageURL
		}

		// Keterangan: try to decrypt again if it looks like it might still be encrypted
		keterangan := res.Keterangan
		if keterangan != "" && h.cryptoSvc != nil {
			// DecryptWithFallback will return the original string if it can't be decrypted
			keterangan = h.cryptoSvc.DecryptWithFallback(keterangan)
		}

		key := res.KawasanID
		if key == "" {
			key = res.Kawasan
		}

		payload.Items = append(payload.Items, map[string]interface{}{
			"no":                      i + 1,
			"pic":                     res.PIC,
			"picName":                 res.PIC,
			"pic_name":                res.PIC,
			"kawasanName":             res.Kawasan,
			"detailKawasanName":       res.DetailKawasan,
			"aspekName":               res.Aspek,
			"detailAspekName":         res.Detail,
			"uraianID":                res.UraianID,
			"uraianName":              res.Uraian,
			"uraian":                  res.Uraian,
			"nilai":                   res.Nilai,
			"total_nilai_perkawasan":  kawasanNilai[key],
			"total_temuan_perkawasan": kawasanTemuan[key],
			"total_nilai_peraspek":    kawasanNilai[key],
			"total_temuan_peraspek":   kawasanTemuan[key],
			"keterangan":              keterangan,
			"dueDate":                 dueDateStr,
			"due_date":                dueDateStr,
			"DueDate":                 dueDateStr,
			"followUp":                followUpStr,
			"follow_up":               followUpStr,
			"follow_up_date":          followUpStr,
			"FollowUpDate":            followUpStr,
			"imageUrl":                imgUrl,
			"image_url":               imgUrl,
		})
	}

		buf, err := exporter.GenerateExcelWithPlaceholder("./templates/master_gmp.xlsx", "Rev 00", payload)
		if err != nil {
			h.log.Error("failed to generate excel with placeholder", logger.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "failed to generate excel file",
			})
		}
	
	fileNameArea := areaName
	if fileNameArea == "Semua Area" {
		fileNameArea = "Semua_Area"
	}

	fileName := fmt.Sprintf("Report_%s_%s.xlsx", fileNameArea, time.Now().Format("20060102_150405"))

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	return c.SendStream(buf)
}

// UploadTemplate accepts a multipart/form-data .xlsx file and overwrites the corporate template.
// POST /api/v1/dashboard/export/template
func (h *DashboardHandler) UploadTemplate(c *fiber.Ctx) error {
	file, err := c.FormFile("template")
	if err != nil {
		h.log.Error("failed to get template file from form", logger.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "failed to upload template",
		})
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".xlsx" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "only .xlsx files are allowed",
		})
	}

	// Save to templates folder, overwriting the existing report_template.xlsx
	savePath := "templates/report_template.xlsx"
	if err := c.SaveFile(file, savePath); err != nil {
		h.log.Error("failed to save template file", logger.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to save template",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "template uploaded successfully",
	})
}

// GetAuditorDetail returns per-auditor inspection statistics.
// GET /api/v1/dashboard/auditor-detail
func (h *DashboardHandler) GetAuditorDetail(c *fiber.Ctx) error {
	allowedAreas := h.getAllowedAreas(c)

	// Step 1: get all users that have at least one inspection as inspector
	type AuditorRow struct {
		UserID    string `gorm:"column:UserID"`
		FullName  string `gorm:"column:FullName"`
		RoleID    string `gorm:"column:RoleID"`
		Total     int64  `gorm:"column:total"`
		Completed int64  `gorm:"column:completed"`
		Ongoing   int64  `gorm:"column:ongoing"`
		Draft     int64  `gorm:"column:draft"`
		Approved  int64  `gorm:"column:approved"`
	}

	var rows []AuditorRow
	queryStr := `
		SELECT
			u."UserID",
			u."FullName",
			u."RoleID",
			COUNT(ih."InspectionID")                                             AS total,
			SUM(CASE WHEN ih."InspectionHeaderStatus" = 'Completed' THEN 1 ELSE 0 END) AS completed,
			SUM(CASE WHEN ih."InspectionHeaderStatus" = 'Ongoing'   THEN 1 ELSE 0 END) AS ongoing,
			SUM(CASE WHEN ih."InspectionHeaderStatus" = 'Draft'     THEN 1 ELSE 0 END) AS draft,
			SUM(CASE WHEN ih."InspectionHeaderStatus" = 'Approved'  THEN 1 ELSE 0 END) AS approved
		FROM "Users" u
		INNER JOIN "Inspection_Header" ih ON ih."InspectorID" = u."UserID"
	`
	var args []interface{}
	if allowedAreas != nil {
		queryStr += ` WHERE ih."AreaID" IN ? `
		args = append(args, allowedAreas)
	}
	queryStr += ` GROUP BY u."UserID", u."FullName", u."RoleID" ORDER BY total DESC `

	err := h.db.Raw(queryStr, args...).Scan(&rows).Error

	if err != nil {
		h.log.Error("failed to get auditor detail", logger.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch auditor detail",
		})
	}

	items := make([]fiber.Map, 0, len(rows))
	for _, r := range rows {
		completionRate := 0.0
		if r.Total > 0 {
			completionRate = float64(r.Completed+r.Approved) / float64(r.Total) * 100
		}
		items = append(items, fiber.Map{
			"user_id":         r.UserID,
			"full_name":       r.FullName,
			"role_id":         r.RoleID,
			"total":           r.Total,
			"completed":       r.Completed,
			"ongoing":         r.Ongoing,
			"draft":           r.Draft,
			"approved":        r.Approved,
			"completion_rate": completionRate,
		})
	}

	return c.JSON(fiber.Map{
		"message": "success",
		"data":    fiber.Map{"items": items},
	})
}

// GetPICDetail returns per-PIC issue follow-up statistics.
// GET /api/v1/dashboard/pic-detail
func (h *DashboardHandler) GetPICDetail(c *fiber.Ctx) error {
	type PICRow struct {
		UserID     string `gorm:"column:UserID"`
		FullName   string `gorm:"column:FullName"`
		RoleID     string `gorm:"column:RoleID"`
		Total      int64  `gorm:"column:total"`
		Open       int64  `gorm:"column:open_count"`
		InProgress int64  `gorm:"column:in_progress"`
		Closed     int64  `gorm:"column:closed_count"`
		Verified   int64  `gorm:"column:verified_count"`
		Overdue    int64  `gorm:"column:overdue_count"`
	}

	var rows []PICRow
	now := time.Now()
	allowedAreas := h.getAllowedAreas(c)

	queryStr := `
		SELECT
			u."UserID",
			u."FullName",
			u."RoleID",
			COUNT(i."IssueID")                                                        AS total,
			SUM(CASE WHEN i."IssueStatus" = 'Open'              THEN 1 ELSE 0 END)     AS open_count,
			SUM(CASE WHEN i."IssueStatus" = 'InProgress'        THEN 1 ELSE 0 END)     AS in_progress,
			SUM(CASE WHEN i."IssueStatus" = 'Closed'            THEN 1 ELSE 0 END)     AS closed_count,
			SUM(CASE WHEN i."IssueStatus" = 'Verified'          THEN 1 ELSE 0 END)     AS verified_count,
			SUM(CASE WHEN i."DueDate" IS NOT NULL AND i."DueDate" < ? AND i."IssueStatus" NOT IN ('Closed','Verified') THEN 1 ELSE 0 END) AS overdue_count
		FROM "Users" u
		INNER JOIN "Issue" i ON i."IssuePICUserID" = u."UserID"
		INNER JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"
		INNER JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"
	`
	args := []interface{}{now}
	if allowedAreas != nil {
		queryStr += ` WHERE ih."AreaID" IN ? `
		args = append(args, allowedAreas)
	}
	queryStr += ` GROUP BY u."UserID", u."FullName", u."RoleID" ORDER BY total DESC `

	err := h.db.Raw(queryStr, args...).Scan(&rows).Error

	if err != nil {
		h.log.Error("failed to get PIC detail", logger.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch PIC detail",
		})
	}

	items := make([]fiber.Map, 0, len(rows))
	for _, r := range rows {
		completionRate := 0.0
		if r.Total > 0 {
			completionRate = float64(r.Closed+r.Verified) / float64(r.Total) * 100
		}
		items = append(items, fiber.Map{
			"user_id":         r.UserID,
			"full_name":       r.FullName,
			"role_id":         r.RoleID,
			"total":           r.Total,
			"open":            r.Open,
			"in_progress":     r.InProgress,
			"closed":          r.Closed,
			"verified":        r.Verified,
			"overdue":         r.Overdue,
			"completion_rate": completionRate,
		})
	}

	return c.JSON(fiber.Map{
		"message": "success",
		"data":    fiber.Map{"items": items},
	})
}


// getAllowedAreas returns a slice of AreaIDs the user is allowed to access.
// If the user is Admin (ROLE-001), it returns nil (meaning no restrictions).
func (h *DashboardHandler) getAllowedAreas(c *fiber.Ctx) []string {
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	roleID := middleware.GetRoleID(c)
	if isSuperAdmin || roleID == "ROLE-000" || roleID == "SUPERADMIN" {
		return nil // SuperAdmin has unrestricted global access across all plants
	}

	userPlantID, _ := c.Locals("userPlantID").(string)
	var areaIDs []string

	if userPlantID != "" {
		// Non-SuperAdmin: Strictly scoped to areas belonging to their assigned PlantID
		h.db.Table("\"Area_Master\"").Select("\"AreaID\"").Where("\"PlantID\" = ?", userPlantID).Scan(&areaIDs)
		if len(areaIDs) == 0 {
			return []string{"RESTRICTED_NONE"}
		}
		return areaIDs
	}

	// Fallback to PIC_Mapping for non-superadmin users without explicit plant assignment
	h.db.Table("\"PIC_Mapping\"").Select("\"AreaID\"").Where("\"UserID\" = ?", middleware.GetUserID(c)).Scan(&areaIDs)
	if len(areaIDs) == 0 {
		return []string{"RESTRICTED_NONE"}
	}
	return areaIDs
}

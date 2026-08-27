package dashboardhandler

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	dashboarddomain "github.com/monitoring-system/backend/internal/domain/dashboard"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/exporter"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/response"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db         *gorm.DB
	log        *logger.Logger
	cryptoSvc  *crypto.Service
	layoutRepo dashboarddomain.DashboardLayoutRepository
}

type issueInitialPhoto struct {
	IssueID  string `gorm:"column:issue_id"`
	ImageURL string `gorm:"column:image_url"`
}

// loadInitialIssueImages returns every initial finding photo grouped by issue.
// GMP rows represent an uraian/result, so follow-up and WO/WR photos must not
// replace the original evidence shown for that uraian.
func (h *DashboardHandler) loadInitialIssueImages(issueIDs []string) (map[string][]string, error) {
	grouped := make(map[string][]string)
	if len(issueIDs) == 0 {
		return grouped, nil
	}

	var photos []issueInitialPhoto
	err := h.db.Table(`"Issue_Photo"`).
		Select(`"IssueID" as issue_id, "ImageUrl" as image_url`).
		Where(`"IssueID" IN ? AND "PhotoType" = ?`, issueIDs, "Initial").
		Order(`"IssueID" ASC, "PhotoCreatedAt" ASC`).
		Scan(&photos).Error
	if err != nil {
		return nil, err
	}

	for _, photo := range photos {
		if photo.ImageURL == "" {
			continue
		}
		imageURL := photo.ImageURL
		if h.cryptoSvc != nil {
			imageURL = h.cryptoSvc.DecryptWithFallback(imageURL)
		}
		grouped[photo.IssueID] = append(grouped[photo.IssueID], imageURL)
	}

	return grouped, nil
}

func NewDashboardHandler(db *gorm.DB, log *logger.Logger, cryptoSvc *crypto.Service, layoutRepo dashboarddomain.DashboardLayoutRepository) *DashboardHandler {
	return &DashboardHandler{
		db:         db,
		log:        log,
		cryptoSvc:  cryptoSvc,
		layoutRepo: layoutRepo,
	}
}

// GetLayout returns the authenticated user's saved dashboard widget layout,
// falling back to the role's default arrangement if they've never customized
// it (or if what they saved fails to parse — never hard-fail the dashboard
// over a preference blob).
func (h *DashboardHandler) GetLayout(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	roleID := middleware.GetRoleID(c)
	if userID == "" {
		return response.Unauthorized(c, "Unauthorized")
	}

	fallback := dashboarddomain.DefaultLayoutForRole(roleID)

	saved, err := h.layoutRepo.FindByUserID(userID)
	if err != nil {
		h.log.Error("dashboard: failed to load layout", logger.Error(err))
		return response.OK(c, "Using default layout", fallback)
	}
	if saved == nil {
		return response.OK(c, "Using default layout", fallback)
	}

	var widgets []dashboarddomain.WidgetConfig
	if err := json.Unmarshal([]byte(saved.LayoutJSON), &widgets); err != nil {
		h.log.Error("dashboard: saved layout is not valid JSON, falling back to default", logger.Error(err))
		return response.OK(c, "Using default layout", fallback)
	}

	return response.OK(c, "Layout fetched", widgets)
}

// SaveLayout persists the authenticated user's widget arrangement (visible +
// order for Fase 3a; x/y/w/h will start arriving once Fase 3b's
// react-grid-layout UI ships).
func (h *DashboardHandler) SaveLayout(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Unauthorized")
	}

	var widgets []dashboarddomain.WidgetConfig
	if err := c.BodyParser(&widgets); err != nil {
		return response.BadRequest(c, "Invalid request body", err.Error())
	}
	if len(widgets) == 0 {
		return response.BadRequest(c, "Layout must contain at least one widget", nil)
	}

	raw, err := json.Marshal(widgets)
	if err != nil {
		return response.InternalServerError(c, "Failed to encode layout", err.Error())
	}

	if err := h.layoutRepo.Upsert(userID, string(raw)); err != nil {
		return response.InternalServerError(c, "Failed to save layout", err.Error())
	}

	return response.OK(c, "Layout saved", widgets)
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
	if allowedAreas != nil {
		qRunning = qRunning.Where("\"AreaID\" IN ?", allowedAreas)
	}
	qRunning.Count(&totalInspectionsRunning)

	qCompleted := h.db.Model(&inspection.InspectionHeader{}).Where("\"InspectionHeaderStatus\" = ?", "Completed")
	if allowedAreas != nil {
		qCompleted = qCompleted.Where("\"AreaID\" IN ?", allowedAreas)
	}
	qCompleted.Count(&inspectionsCompleted)

	// Issue queries need join with Inspection_Header
	qIssueBase := h.db.Table("\"Issue\" i").Joins("JOIN \"Inspection_Result\" ir ON ir.\"ResultID\" = i.\"ResultID\"").Joins("JOIN \"Inspection_Header\" ih ON ih.\"InspectionID\" = ir.\"InspectionID\"")
	qIssueOpen := qIssueBase.Session(&gorm.Session{}).Where("i.\"IssueStatus\" NOT IN ('Closed', 'Verified', 'ClosedOverdue')")
	if allowedAreas != nil {
		qIssueOpen = qIssueOpen.Where("ih.\"AreaID\" IN ?", allowedAreas)
	}
	qIssueOpen.Count(&totalOpenIssues)

	qIssueClosed := qIssueBase.Session(&gorm.Session{}).Where("i.\"IssueStatus\" IN ('Closed', 'Verified', 'ClosedOverdue')")
	if allowedAreas != nil {
		qIssueClosed = qIssueClosed.Where("ih.\"AreaID\" IN ?", allowedAreas)
	}
	qIssueClosed.Count(&picFollowupCompleted)

	qIssueOverdue := qIssueBase.Session(&gorm.Session{}).Where("i.\"IssueStatus\" NOT IN ('Closed', 'Verified', 'ClosedOverdue') AND i.\"DueDate\" < NOW()")
	if allowedAreas != nil {
		qIssueOverdue = qIssueOverdue.Where("ih.\"AreaID\" IN ?", allowedAreas)
	}
	qIssueOverdue.Count(&picFollowupOverdue)

	qResultBase := h.db.Table("\"Inspection_Result\" ir").Joins("JOIN \"Inspection_Header\" ih ON ih.\"InspectionID\" = ir.\"InspectionID\"")
	qCheck := qResultBase.Session(&gorm.Session{})
	if allowedAreas != nil {
		qCheck = qCheck.Where("ih.\"AreaID\" IN ?", allowedAreas)
	}
	qCheck.Count(&totalCheck)

	qOK := qResultBase.Session(&gorm.Session{}).Where("ir.\"Checking\" = ?", "OK")
	if allowedAreas != nil {
		qOK = qOK.Where("ih.\"AreaID\" IN ?", allowedAreas)
	}
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
	queryStr += ` GROUP BY am."AreaID", am."AreaName" ORDER BY open_issues DESC `
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

			mRate := 0.0
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

			mRate := 0.0
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

			mRate := 0.0
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

			mRate := 0.0
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

	// Calculate WOWR Statistics
	var wowrTotal, wowrVerified, wowrPending, wowrRejected, wowrAwaiting int64
	qWOWRBase := h.db.Table("\"Issue\" i").
		Joins("JOIN \"Inspection_Result\" ir ON ir.\"ResultID\" = i.\"ResultID\"").
		Joins("JOIN \"Inspection_Header\" ih ON ih.\"InspectionID\" = ir.\"InspectionID\"").
		Where("i.\"NeedsWOWR\" = true OR (i.\"WO_ID\" IS NOT NULL AND i.\"WO_ID\" != '') OR (i.\"WR_ID\" IS NOT NULL AND i.\"WR_ID\" != '')")
	if allowedAreas != nil {
		qWOWRBase = qWOWRBase.Where("ih.\"AreaID\" IN ?", allowedAreas)
	}
	qWOWRBase.Session(&gorm.Session{}).Count(&wowrTotal)
	qWOWRBase.Session(&gorm.Session{}).Where("i.\"WOWRStatus\" = ?", "Verified").Count(&wowrVerified)
	qWOWRBase.Session(&gorm.Session{}).Where("i.\"WOWRStatus\" = ?", "PendingValidation").Count(&wowrPending)
	qWOWRBase.Session(&gorm.Session{}).Where("i.\"WOWRStatus\" = ?", "Rejected").Count(&wowrRejected)
	qWOWRBase.Session(&gorm.Session{}).Where("i.\"WOWRStatus\" = ? OR i.\"WOWRStatus\" IS NULL", "None").Count(&wowrAwaiting)

	wowrVerifiedRate := 0.0
	if wowrTotal > 0 {
		wowrVerifiedRate = float64(wowrVerified) / float64(wowrTotal) * 100
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
		"issues_resolved":           picFollowupCompleted,
		"pic_followup_overdue":      picFollowupOverdue,
		"auditee_status":            auditeeStatusList,
		"compliance_trend":          complianceTrend,
		"wowr_total":                wowrTotal,
		"wowr_verified":             wowrVerified,
		"wowr_pending":              wowrPending,
		"wowr_rejected":             wowrRejected,
		"wowr_awaiting":             wowrAwaiting,
		"wowr_verified_rate":        float64(int(wowrVerifiedRate*10)) / 10.0,
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
		ImageURLs       []string   `gorm:"-" json:"image_urls"`
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

	issueIDs := make([]string, 0)
	for _, result := range results {
		if result.IssueID != nil && *result.IssueID != "" {
			issueIDs = append(issueIDs, *result.IssueID)
		}
	}
	initialImages, err := h.loadInitialIssueImages(issueIDs)
	if err != nil {
		h.log.Error("Failed to fetch GMP initial issue photos", logger.Error(err))
		return response.InternalServerError(c, "Failed to load issue photos", err.Error())
	}

	// TotalNilai remains an aggregate per kawasan. TotalTemuan is intentionally
	// per uraian so the same kawasan count is not repeated on every row.
	kawasanNilai := make(map[string]int)

	for _, res := range results {
		key := res.KawasanID
		if key == "" {
			key = res.Kawasan
		}
		kawasanNilai[key] += res.Nilai
	}

	for i := range results {
		key := results[i].KawasanID
		if key == "" {
			key = results[i].Kawasan
		}
		results[i].TotalNilai = kawasanNilai[key]
		results[i].TotalTemuan = 0
		if results[i].IssueID != nil && *results[i].IssueID != "" {
			results[i].TotalTemuan = 1
			results[i].ImageURLs = initialImages[*results[i].IssueID]
			if len(results[i].ImageURLs) > 0 {
				firstImage := results[i].ImageURLs[0]
				results[i].ImageURL = &firstImage
			}
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
		ImageURLs       []string   `gorm:"-" json:"image_urls"`
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
	issueIDs := make([]string, 0)
	for _, result := range results {
		if result.IssueID != nil && *result.IssueID != "" {
			issueIDs = append(issueIDs, *result.IssueID)
		}
	}
	initialImages, err := h.loadInitialIssueImages(issueIDs)
	if err != nil {
		h.log.Error("Failed to fetch GMP export issue photos", logger.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to fetch issue photos",
		})
	}

	for i := range results {
		totalNilai += results[i].Nilai
		key := results[i].KawasanID
		if key == "" {
			key = results[i].Kawasan
		}
		kawasanNilai[key] += results[i].Nilai
		if results[i].IssueID != nil && *results[i].IssueID != "" {
			totalTemuan++
			results[i].ImageURLs = initialImages[*results[i].IssueID]
			if len(results[i].ImageURLs) > 0 {
				firstImage := results[i].ImageURLs[0]
				results[i].ImageURL = &firstImage
			}
		}

		if h.cryptoSvc != nil {
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
			"datetime.now()":     time.Now().Format("02-Jan-06"),
			"AreaName":           areaName,
			"PICName":            picName,
			"DueDate":            headerDueDate,
			"total_semua_nilai":  totalNilai,
			"total_semua_temuan": totalTemuan,
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
		rowTemuan := 0
		if res.IssueID != nil && *res.IssueID != "" {
			rowTemuan = 1
		}
		var imageValue interface{} = ""
		if len(res.ImageURLs) > 0 {
			imageValue = res.ImageURLs
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
			"total_temuan_perkawasan": rowTemuan,
			"total_nilai_peraspek":    kawasanNilai[key],
			"total_temuan_peraspek":   rowTemuan,
			"keterangan":              keterangan,
			"dueDate":                 dueDateStr,
			"due_date":                dueDateStr,
			"DueDate":                 dueDateStr,
			"followUp":                followUpStr,
			"follow_up":               followUpStr,
			"follow_up_date":          followUpStr,
			"FollowUpDate":            followUpStr,
			"imageUrl":                imageValue,
			"image_url":               imageValue,
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
// SuperAdmin can view all plants (nil) or filter by query parameter `plant_id`.
// Non-SuperAdmin users (Admin Plant, Auditee) are strictly scoped to their assigned `userPlantID`.
func (h *DashboardHandler) getAllowedAreas(c *fiber.Ctx) []string {
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	roleID := middleware.GetRoleID(c)
	if roleID == "ROLE-000" || roleID == "SUPERADMIN" {
		isSuperAdmin = true
	}

	queryPlantID := c.Query("plant_id")

	// Super Admin Access Logic
	if isSuperAdmin {
		if queryPlantID == "" || queryPlantID == "all" {
			return nil // Unrestricted access to all areas across all plants
		}

		var areaIDs []string
		if queryPlantID == "GLOBAL" || queryPlantID == "NULL" {
			h.db.Table("\"Area_Master\"").Select("\"AreaID\"").Where("\"PlantID\" IS NULL OR \"PlantID\" = ''").Scan(&areaIDs)
		} else {
			h.db.Table("\"Area_Master\"").Select("\"AreaID\"").Where("\"PlantID\" = ?", queryPlantID).Scan(&areaIDs)
		}
		if len(areaIDs) == 0 {
			return []string{"RESTRICTED_NONE"}
		}
		return areaIDs
	}

	// Non-SuperAdmin Access Logic (Admin Plant, Auditee, Auditor with plant restriction)
	userPlantID, _ := c.Locals("userPlantID").(string)

	if userPlantID != "" {
		var areaIDs []string
		if userPlantID == "GLOBAL" || userPlantID == "NULL" {
			h.db.Table("\"Area_Master\"").Select("\"AreaID\"").Where("\"PlantID\" IS NULL OR \"PlantID\" = ''").Scan(&areaIDs)
		} else {
			h.db.Table("\"Area_Master\"").Select("\"AreaID\"").Where("\"PlantID\" = ?", userPlantID).Scan(&areaIDs)
		}
		if len(areaIDs) == 0 {
			return []string{"RESTRICTED_NONE"}
		}
		return areaIDs
	}

	// For Auditor role without specific plant assignment, return nil (all areas)
	if middleware.IsAuditorRole(roleID) {
		return nil
	}

	// Fallback to PIC_Mapping for Auditees
	var areaIDs []string
	h.db.Table("\"PIC_Mapping\"").Select("\"AreaID\"").Where("\"UserID\" = ?", middleware.GetUserID(c)).Scan(&areaIDs)
	if len(areaIDs) == 0 {
		return []string{"RESTRICTED_NONE"}
	}
	return areaIDs
}

// GetWOWRReport returns detailed WOWR analytics summary, rates, area breakdown, and list items.
// GET /api/v1/dashboard/wowr-report
func (h *DashboardHandler) GetWOWRReport(c *fiber.Ctx) error {
	areaID := c.Query("area_id")
	kawasanID := c.Query("kawasan_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	allowedAreas := h.getAllowedAreas(c)

	type WOWRReportItem struct {
		IssueID            string     `json:"issue_id"`
		PhotoID            string     `json:"photo_id,omitempty"`
		WO_ID              string     `json:"wo_id"`
		WR_ID              string     `json:"wr_id"`
		NeedsWOWR          bool       `json:"needs_wo_wr"`
		WOWRStatus         string     `json:"wowr_status"`
		IssueStatus        string     `json:"issue_status"`
		AreaID             string     `json:"area_id"`
		AreaName           string     `json:"area_name"`
		KawasanID          string     `json:"kawasan_id"`
		KawasanName        string     `json:"kawasan_name"`
		DetailKawasanID    string     `json:"detail_kawasan_id"`
		DetailKawasanName  string     `json:"detail_kawasan_name"`
		PICName            string     `json:"pic_name"`
		AspekName          string     `json:"aspek_name"`
		DetailAspekName    string     `json:"detail_aspek_name"`
		UraianText         string     `json:"uraian_text"`
		HabitName          string     `json:"habit_name"`
		EquipmentName      string     `json:"equipment_name"`
		InfrastructureName string     `json:"infrastructure_name"`
		Keterangan         string     `json:"keterangan"`
		DueDate            *time.Time `json:"due_date"`
		CreatedAt          time.Time  `json:"created_at"`
	}

	issueQuery := h.db.Table(`"Issue" i`).
		Select(`i."IssueID" as issue_id,
			'' as photo_id,
			COALESCE(i."WO_ID", '') as wo_id,
			COALESCE(i."WR_ID", '') as wr_id,
			i."NeedsWOWR" as needs_wo_wr,
			COALESCE(NULLIF(i."WOWRStatus", ''), 'None') as wowr_status,
			i."IssueStatus" as issue_status,
			ih."AreaID" as area_id,
			COALESCE(am_area."AreaName", ih."AreaID") as area_name,
			ih."KawasanID" as kawasan_id,
			COALESCE(km."KawasanName", ih."KawasanID") as kawasan_name,
			ih."DetailKawasanID" as detail_kawasan_id,
			COALESCE(dkm."DetailKawasanName", ih."DetailKawasanID") as detail_kawasan_name,
			COALESCE(u_pic."FullName", i."IssuePICUserID") as pic_name,
			COALESCE(asp."AspekName", '') as aspek_name,
			COALESCE(dm."DetailName", '') as detail_aspek_name,
			COALESCE(um."UraianText", '') as uraian_text,
			'' as habit_name,
			'' as equipment_name,
			'' as infrastructure_name,
			i."Keterangan" as keterangan,
			i."DueDate" as due_date,
			i."IssueCreatedAt" as created_at`).
		Joins(`JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"`).
		Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Joins(`LEFT JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"`).
		Joins(`LEFT JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"`).
		Joins(`LEFT JOIN "Aspek_Master" asp ON asp."AspekID" = dm."AspekID"`).
		Joins(`LEFT JOIN "Area_Master" am_area ON am_area."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u_pic ON u_pic."UserID" = i."IssuePICUserID"`).
		Where(`(i."NeedsWOWR" = true OR COALESCE(i."WO_ID", '') != '' OR COALESCE(i."WR_ID", '') != '')
			AND NOT EXISTS (
				SELECT 1 FROM "Issue_Photo" ip_request
				WHERE ip_request."IssueID" = i."IssueID"
				  AND ip_request."PhotoType" = 'Initial'
				  AND (ip_request."NeedsWOWR" = true OR COALESCE(ip_request."WO_ID", '') != '' OR COALESCE(ip_request."WR_ID", '') != '')
			)`)

	photoQuery := h.db.Table(`"Issue_Photo" ip`).
		Select(`i."IssueID" as issue_id,
			ip."IssuePhotoID" as photo_id,
			COALESCE(ip."WO_ID", '') as wo_id,
			COALESCE(ip."WR_ID", '') as wr_id,
			ip."NeedsWOWR" as needs_wo_wr,
			COALESCE(NULLIF(ip."WOWRStatus", ''), 'None') as wowr_status,
			i."IssueStatus" as issue_status,
			ih."AreaID" as area_id,
			COALESCE(am_area."AreaName", ih."AreaID") as area_name,
			ih."KawasanID" as kawasan_id,
			COALESCE(km."KawasanName", ih."KawasanID") as kawasan_name,
			ih."DetailKawasanID" as detail_kawasan_id,
			COALESCE(dkm."DetailKawasanName", ih."DetailKawasanID") as detail_kawasan_name,
			COALESCE(u_pic."FullName", i."IssuePICUserID") as pic_name,
			COALESCE(asp."AspekName", '') as aspek_name,
			COALESCE(dm."DetailName", '') as detail_aspek_name,
			COALESCE(um."UraianText", '') as uraian_text,
			CASE WHEN LOWER(COALESCE(hm."CategoryName", ip."HEICategory", '')) = 'habit' THEN COALESCE(hm."HEIName", '') ELSE '' END as habit_name,
			CASE WHEN LOWER(COALESCE(hm."CategoryName", ip."HEICategory", '')) = 'equipment' THEN COALESCE(hm."HEIName", '') ELSE '' END as equipment_name,
			CASE WHEN LOWER(COALESCE(hm."CategoryName", ip."HEICategory", '')) = 'infrastructure' THEN COALESCE(hm."HEIName", '') ELSE '' END as infrastructure_name,
			COALESCE(NULLIF(ip."Keterangan", ''), i."Keterangan") as keterangan,
			i."DueDate" as due_date,
			ip."PhotoUpdatedAt" as created_at`).
		Joins(`JOIN "Issue" i ON i."IssueID" = ip."IssueID"`).
		Joins(`JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"`).
		Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Joins(`LEFT JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"`).
		Joins(`LEFT JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"`).
		Joins(`LEFT JOIN "Aspek_Master" asp ON asp."AspekID" = dm."AspekID"`).
		Joins(`LEFT JOIN "HEI_Master" hm ON hm."HEIID" = ip."HEIID"`).
		Joins(`LEFT JOIN "Area_Master" am_area ON am_area."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u_pic ON u_pic."UserID" = i."IssuePICUserID"`).
		Where(`ip."PhotoType" = 'Initial'
			AND (ip."NeedsWOWR" = true OR COALESCE(ip."WO_ID", '') != '' OR COALESCE(ip."WR_ID", '') != '')`)

	if areaID != "" {
		issueQuery = issueQuery.Where(`ih."AreaID" = ?`, areaID)
		photoQuery = photoQuery.Where(`ih."AreaID" = ?`, areaID)
	}
	if kawasanID != "" {
		issueQuery = issueQuery.Where(`ih."KawasanID" = ?`, kawasanID)
		photoQuery = photoQuery.Where(`ih."KawasanID" = ?`, kawasanID)
	}
	if startDate != "" {
		issueQuery = issueQuery.Where(`i."IssueCreatedAt" >= ?`, startDate+" 00:00:00")
		photoQuery = photoQuery.Where(`ip."PhotoUpdatedAt" >= ?`, startDate+" 00:00:00")
	}
	if endDate != "" {
		issueQuery = issueQuery.Where(`i."IssueCreatedAt" <= ?`, endDate+" 23:59:59")
		photoQuery = photoQuery.Where(`ip."PhotoUpdatedAt" <= ?`, endDate+" 23:59:59")
	}
	if allowedAreas != nil {
		issueQuery = issueQuery.Where(`ih."AreaID" IN ?`, allowedAreas)
		photoQuery = photoQuery.Where(`ih."AreaID" IN ?`, allowedAreas)
	}

	var items []WOWRReportItem
	if err := issueQuery.Order(`i."IssueCreatedAt" DESC`).Scan(&items).Error; err != nil {
		h.log.Error("Failed to fetch WOWR report items", logger.Error(err))
		return response.InternalServerError(c, "Failed to load WOWR report data", err.Error())
	}
	var photoItems []WOWRReportItem
	if err := photoQuery.Order(`ip."PhotoUpdatedAt" DESC`).Scan(&photoItems).Error; err != nil {
		h.log.Error("Failed to fetch photo WOWR report items", logger.Error(err))
		return response.InternalServerError(c, "Failed to load photo WOWR report data", err.Error())
	}
	items = append(photoItems, items...)

	var total, verified, pending, rejected, awaiting int64
	total = int64(len(items))

	type AreaBreakdown struct {
		AreaName string `json:"area_name"`
		Total    int64  `json:"total"`
		Verified int64  `json:"verified"`
		Pending  int64  `json:"pending"`
		Rejected int64  `json:"rejected"`
		Awaiting int64  `json:"awaiting"`
	}
	areaMap := make(map[string]*AreaBreakdown)

	for i := range items {
		if h.cryptoSvc != nil && items[i].Keterangan != "" {
			items[i].Keterangan = h.cryptoSvc.DecryptWithFallback(items[i].Keterangan)
		}

		st := items[i].WOWRStatus
		if st == "Verified" {
			verified++
		} else if st == "PendingValidation" {
			pending++
		} else if st == "Rejected" {
			rejected++
		} else {
			awaiting++
		}

		an := items[i].AreaName
		if an == "" {
			an = "Lainnya"
		}
		if _, ok := areaMap[an]; !ok {
			areaMap[an] = &AreaBreakdown{AreaName: an}
		}
		ab := areaMap[an]
		ab.Total++
		if st == "Verified" {
			ab.Verified++
		} else if st == "PendingValidation" {
			ab.Pending++
		} else if st == "Rejected" {
			ab.Rejected++
		} else {
			ab.Awaiting++
		}
	}

	var byArea []AreaBreakdown
	for _, ab := range areaMap {
		byArea = append(byArea, *ab)
	}

	verifiedRate := 0.0
	pendingRate := 0.0
	rejectedRate := 0.0
	awaitingRate := 0.0

	if total > 0 {
		verifiedRate = float64(verified) / float64(total) * 100
		pendingRate = float64(pending) / float64(total) * 100
		rejectedRate = float64(rejected) / float64(total) * 100
		awaitingRate = float64(awaiting) / float64(total) * 100
	}

	return response.OK(c, "success", fiber.Map{
		"summary": fiber.Map{
			"total":         total,
			"verified":      verified,
			"pending":       pending,
			"rejected":      rejected,
			"awaiting":      awaiting,
			"verified_rate": float64(int(verifiedRate*10)) / 10.0,
			"pending_rate":  float64(int(pendingRate*10)) / 10.0,
			"rejected_rate": float64(int(rejectedRate*10)) / 10.0,
			"awaiting_rate": float64(int(awaitingRate*10)) / 10.0,
		},
		"by_area": byArea,
		"items":   items,
	})
}

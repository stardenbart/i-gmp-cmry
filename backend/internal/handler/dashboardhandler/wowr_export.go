package dashboardhandler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/exporter"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/response"
	"gorm.io/gorm"
)

type wowrExportItem struct {
	IssueID            string     `gorm:"column:issue_id"`
	PhotoID            string     `gorm:"column:photo_id"`
	WOID               string     `gorm:"column:wo_id"`
	WRID               string     `gorm:"column:wr_id"`
	WOWRStatus         string     `gorm:"column:wowr_status"`
	IssueStatus        string     `gorm:"column:issue_status"`
	AreaName           string     `gorm:"column:area_name"`
	KawasanName        string     `gorm:"column:kawasan_name"`
	DetailKawasanName  string     `gorm:"column:detail_kawasan_name"`
	PICName            string     `gorm:"column:pic_name"`
	AspekName          string     `gorm:"column:aspek_name"`
	DetailAspekName    string     `gorm:"column:detail_aspek_name"`
	UraianText         string     `gorm:"column:uraian_text"`
	HabitName          string     `gorm:"column:habit_name"`
	EquipmentName      string     `gorm:"column:equipment_name"`
	InfrastructureName string     `gorm:"column:infrastructure_name"`
	Keterangan         string     `gorm:"column:keterangan"`
	DueDate            *time.Time `gorm:"column:due_date"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
}

func wowrReportSelect(issueLevel bool) string {
	photoID := `ip."IssuePhotoID"`
	woID := `ip."WO_ID"`
	wrID := `ip."WR_ID"`
	wowrStatus := `ip."WOWRStatus"`
	keterangan := `COALESCE(NULLIF(ip."Keterangan", ''), i."Keterangan")`
	createdAt := `ip."PhotoUpdatedAt"`
	habit := `CASE WHEN LOWER(COALESCE(hm."CategoryName", ip."HEICategory", '')) = 'habit' THEN COALESCE(hm."HEIName", '') ELSE '' END`
	equipment := `CASE WHEN LOWER(COALESCE(hm."CategoryName", ip."HEICategory", '')) = 'equipment' THEN COALESCE(hm."HEIName", '') ELSE '' END`
	infrastructure := `CASE WHEN LOWER(COALESCE(hm."CategoryName", ip."HEICategory", '')) = 'infrastructure' THEN COALESCE(hm."HEIName", '') ELSE '' END`
	if issueLevel {
		photoID = `''`
		woID = `i."WO_ID"`
		wrID = `i."WR_ID"`
		wowrStatus = `i."WOWRStatus"`
		keterangan = `i."Keterangan"`
		createdAt = `i."IssueCreatedAt"`
		habit, equipment, infrastructure = `''`, `''`, `''`
	}

	return fmt.Sprintf(`i."IssueID" as issue_id,
		%s as photo_id,
		COALESCE(%s, '') as wo_id,
		COALESCE(%s, '') as wr_id,
		COALESCE(NULLIF(%s, ''), 'None') as wowr_status,
		i."IssueStatus" as issue_status,
		COALESCE(am_area."AreaName", ih."AreaID") as area_name,
		COALESCE(km."KawasanName", ih."KawasanID") as kawasan_name,
		COALESCE(dkm."DetailKawasanName", ih."DetailKawasanID") as detail_kawasan_name,
		COALESCE(u_pic."FullName", i."IssuePICUserID") as pic_name,
		COALESCE(asp."AspekName", '') as aspek_name,
		COALESCE(dm."DetailName", '') as detail_aspek_name,
		COALESCE(um."UraianText", '') as uraian_text,
		%s as habit_name,
		%s as equipment_name,
		%s as infrastructure_name,
		%s as keterangan,
		i."DueDate" as due_date,
		%s as created_at`, photoID, woID, wrID, wowrStatus, habit, equipment, infrastructure, keterangan, createdAt)
}

func applyWOWRExportFilters(c *fiber.Ctx, query *gorm.DB, issueLevel bool, allowedAreas []string) *gorm.DB {
	areaID := c.Query("area_id")
	kawasanID := c.Query("kawasan_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	dateColumn := `ip."PhotoUpdatedAt"`
	if issueLevel {
		dateColumn = `i."IssueCreatedAt"`
	}
	if areaID != "" {
		query = query.Where(`ih."AreaID" = ?`, areaID)
	}
	if kawasanID != "" {
		query = query.Where(`ih."KawasanID" = ?`, kawasanID)
	}
	if startDate != "" {
		query = query.Where(dateColumn+` >= ?`, startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where(dateColumn+` <= ?`, endDate+" 23:59:59")
	}
	if allowedAreas != nil {
		query = query.Where(`ih."AreaID" IN ?`, allowedAreas)
	}
	return query
}

func (h *DashboardHandler) loadWOWRExportItems(c *fiber.Ctx) ([]wowrExportItem, error) {
	joins := func(query *gorm.DB, includeHEI bool) *gorm.DB {
		query = query.
			Joins(`JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"`).
			Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
			Joins(`LEFT JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"`).
			Joins(`LEFT JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"`).
			Joins(`LEFT JOIN "Aspek_Master" asp ON asp."AspekID" = dm."AspekID"`).
			Joins(`LEFT JOIN "Area_Master" am_area ON am_area."AreaID" = ih."AreaID"`).
			Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
			Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
			Joins(`LEFT JOIN "Users" u_pic ON u_pic."UserID" = i."IssuePICUserID"`)
		if includeHEI {
			query = query.Joins(`LEFT JOIN "HEI_Master" hm ON hm."HEIID" = ip."HEIID"`)
		}
		return query
	}

	allowedAreas := h.getAllowedAreas(c)
	issueQuery := joins(h.db.Table(`"Issue" i`).Select(wowrReportSelect(true)), false).
		Where(`(i."NeedsWOWR" = true OR COALESCE(i."WO_ID", '') != '' OR COALESCE(i."WR_ID", '') != '')
			AND NOT EXISTS (
				SELECT 1 FROM "Issue_Photo" ip_request
				WHERE ip_request."IssueID" = i."IssueID" AND ip_request."PhotoType" = 'Initial'
				AND (ip_request."NeedsWOWR" = true OR COALESCE(ip_request."WO_ID", '') != '' OR COALESCE(ip_request."WR_ID", '') != '')
			)`)
	issueQuery = applyWOWRExportFilters(c, issueQuery, true, allowedAreas)

	photoQuery := joins(h.db.Table(`"Issue_Photo" ip`).
		Select(wowrReportSelect(false)).
		Joins(`JOIN "Issue" i ON i."IssueID" = ip."IssueID"`), true).
		Where(`ip."PhotoType" = 'Initial'
			AND (ip."NeedsWOWR" = true OR COALESCE(ip."WO_ID", '') != '' OR COALESCE(ip."WR_ID", '') != '')`)
	photoQuery = applyWOWRExportFilters(c, photoQuery, false, allowedAreas)

	var issueItems, photoItems []wowrExportItem
	if err := issueQuery.Order(`i."IssueCreatedAt" DESC`).Scan(&issueItems).Error; err != nil {
		return nil, err
	}
	if err := photoQuery.Order(`ip."PhotoUpdatedAt" DESC`).Scan(&photoItems).Error; err != nil {
		return nil, err
	}
	return append(photoItems, issueItems...), nil
}

func evidenceDetails(items []WOWRReportEvidence) string {
	if len(items) == 0 {
		return "Belum ada bukti"
	}
	lines := make([]string, 0, len(items))
	for index, item := range items {
		detail := fmt.Sprintf("%d. %s", index+1, item.CreatedAt.Format("02-Jan-2006 15:04"))
		if item.Uploader != "" {
			detail += " | " + item.Uploader
		}
		if item.Description != "" {
			detail += " | " + item.Description
		}
		lines = append(lines, detail)
	}
	return strings.Join(lines, "\n")
}

func heiDetail(item wowrExportItem) string {
	values := make([]string, 0, 3)
	if item.HabitName != "" {
		values = append(values, "Habit: "+item.HabitName)
	}
	if item.EquipmentName != "" {
		values = append(values, "Equipment: "+item.EquipmentName)
	}
	if item.InfrastructureName != "" {
		values = append(values, "Infrastructure: "+item.InfrastructureName)
	}
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, "\n")
}

func evidenceURLs(items []WOWRReportEvidence) []string {
	urls := make([]string, 0, len(items))
	for _, item := range items {
		if item.ImageURL != "" {
			urls = append(urls, item.ImageURL)
		}
	}
	return urls
}

// ExportWOWRReport generates a detailed, print-ready XLSX report with embedded evidence photos.
func (h *DashboardHandler) ExportWOWRReport(c *fiber.Ctx) error {
	items, err := h.loadWOWRExportItems(c)
	if err != nil {
		h.log.Error("Failed to fetch WOWR export data", logger.Error(err))
		return response.InternalServerError(c, "Gagal memuat data ekspor WO/WR", err.Error())
	}

	areaNameFilter := strings.TrimSpace(c.Query("area_name"))
	search := strings.ToLower(strings.TrimSpace(c.Query("search")))
	filtered := make([]wowrExportItem, 0, len(items))
	for _, item := range items {
		if h.cryptoSvc != nil && item.Keterangan != "" {
			item.Keterangan = h.cryptoSvc.DecryptWithFallback(item.Keterangan)
		}
		if areaNameFilter != "" && areaNameFilter != "ALL" && item.AreaName != areaNameFilter {
			continue
		}
		haystack := strings.ToLower(strings.Join([]string{
			item.IssueID, item.WOID, item.WRID, item.AreaName, item.KawasanName,
			item.DetailKawasanName, item.PICName, item.Keterangan, item.AspekName,
			item.DetailAspekName, item.UraianText,
		}, " "))
		if search != "" && !strings.Contains(haystack, search) {
			continue
		}
		filtered = append(filtered, item)
	}

	issueIDs := make([]string, 0, len(filtered))
	seen := make(map[string]bool)
	for _, item := range filtered {
		if !seen[item.IssueID] {
			seen[item.IssueID] = true
			issueIDs = append(issueIDs, item.IssueID)
		}
	}
	evidenceByIssue, err := h.loadWOWREvidence(issueIDs)
	if err != nil {
		h.log.Error("Failed to fetch WOWR export evidence", logger.Error(err))
		return response.InternalServerError(c, "Gagal memuat foto bukti WO/WR", err.Error())
	}

	verified, pending, rejected, awaiting := 0, 0, 0, 0
	payload := &exporter.PlaceholderPayload{
		Headers: map[string]interface{}{
			"generated_at": time.Now().Format("02 January 2006 15:04 MST"),
			"area_filter":  areaNameFilter,
			"period_filter": func() string {
				start, end := c.Query("start_date"), c.Query("end_date")
				if start == "" && end == "" {
					return "Semua periode"
				}
				return fmt.Sprintf("%s s.d. %s", defaultText(start, "Awal"), defaultText(end, "Sekarang"))
			}(),
		},
		Items: make([]map[string]interface{}, 0, len(filtered)),
	}
	if areaNameFilter == "" || areaNameFilter == "ALL" {
		payload.Headers["area_filter"] = "Semua Area"
	}

	for index, item := range filtered {
		switch item.WOWRStatus {
		case "Verified":
			verified++
		case "PendingValidation":
			pending++
		case "Rejected":
			rejected++
		default:
			awaiting++
		}
		initialEvidence, completionEvidence := evidenceForWOWRItem(evidenceByIssue[item.IssueID], item.PhotoID)
		dueDate := "-"
		if item.DueDate != nil {
			dueDate = item.DueDate.Format("02-Jan-2006")
		}
		payload.Items = append(payload.Items, map[string]interface{}{
			"no":                 index + 1,
			"issue_id":           item.IssueID,
			"photo_id":           defaultText(item.PhotoID, "-"),
			"wo_number":          defaultText(item.WOID, "-"),
			"wr_number":          defaultText(item.WRID, "-"),
			"wowr_status":        item.WOWRStatus,
			"issue_status":       item.IssueStatus,
			"location":           strings.Join(nonEmpty(item.AreaName, item.KawasanName, item.DetailKawasanName), "\n"),
			"pic_name":           defaultText(item.PICName, "-"),
			"finding_detail":     strings.Join(nonEmpty(item.AspekName, item.DetailAspekName, item.UraianText), "\n"),
			"hei_detail":         heiDetail(item),
			"description":        defaultText(item.Keterangan, "-"),
			"due_date":           dueDate,
			"created_at":         item.CreatedAt.Format("02-Jan-2006 15:04"),
			"initial_count":      len(initialEvidence),
			"completion_count":   len(completionEvidence),
			"initial_details":    evidenceDetails(initialEvidence),
			"completion_details": evidenceDetails(completionEvidence),
			"initial_images":     evidenceURLs(initialEvidence),
			"completion_images":  evidenceURLs(completionEvidence),
		})
	}
	payload.Headers["total"] = len(filtered)
	payload.Headers["verified"] = verified
	payload.Headers["pending"] = pending
	payload.Headers["rejected"] = rejected
	payload.Headers["awaiting"] = awaiting
	payload.TotalRow = &exporter.TableTotalRow{
		Label:            "TOTAL KESELURUHAN WO/WR",
		LabelStartColumn: "A",
		LabelEndColumn:   "R",
		ValueColumn:      "S",
		Value:            len(filtered),
	}

	buf, err := exporter.GenerateExcelWithPlaceholder("./templates/wowr_report.xlsx", "Laporan WO-WR", payload)
	if err != nil {
		h.log.Error("Failed to generate WOWR Excel report", logger.Error(err))
		return response.InternalServerError(c, "Gagal membuat file Excel WO/WR", err.Error())
	}

	fileName := fmt.Sprintf("Laporan_WO_WR_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	return c.SendStream(buf)
}

func defaultText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func nonEmpty(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

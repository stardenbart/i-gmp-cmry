package issuehandler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/exporter"
	"github.com/monitoring-system/backend/pkg/response"
)

const issueExportLimit = 10000

// IssueFilterHandler handles GET /api/v1/issues/filter.
type IssueFilterHandler struct {
	uc issue.IssueFilterUseCase
}

// NewIssueFilterHandler creates a new IssueFilterHandler.
func NewIssueFilterHandler(uc issue.IssueFilterUseCase) *IssueFilterHandler {
	return &IssueFilterHandler{uc: uc}
}

// GetFiltered handles the filter + facet request for issues.
// @Summary      Filter issues with facets
// @Description  Returns paginated issues with facet counts. Auditee can only see their own.
// @Tags         issues
// @Produce      json
// @Param        q                   query  string  false  "Free-text search on Label"
// @Param        status              query  string  false  "Exact status"
// @Param        status__in          query  string  false  "Comma-separated statuses (OR)"
// @Param        issue_pic_user_id   query  string  false  "Filter by PIC user"
// @Param        wowr_status         query  string  false  "Filter by WOWR status"
// @Param        needs_wo_wr         query  string  false  "true/false"
// @Param        date_from           query  string  false  "Created from"
// @Param        date_to             query  string  false  "Created to"
// @Param        due_from            query  string  false  "DueDate from"
// @Param        due_to              query  string  false  "DueDate to"
// @Param        sort_by             query  string  false  "created_at|due_date|status"
// @Param        sort_order          query  string  false  "asc or desc"
// @Param        page                query  int     false  "Page number"
// @Param        limit               query  int     false  "Items per page"
// @Success      200  {object}  response.APIResponse
// @Failure      500  {object}  response.APIResponse
// @Router       /api/v1/issues/filter [get]
// @Security     BearerAuth
func (h *IssueFilterHandler) GetFiltered(c *fiber.Ctx) error {
	f, err := issue.NewIssueFilter(c)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	applyIssueRequestScope(c, f)

	result, err := h.uc.GetFiltered(f)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch issues", err.Error())
	}

	return response.OK(c, "success", result)
}

func applyIssueRequestScope(c *fiber.Ctx, f *issue.IssueFilter) {
	// Apply role-based scope: auditee can only see their own issues
	actorRole := middleware.GetRoleID(c)
	actorID := middleware.GetUserID(c)
	if !middleware.IsAuditorRole(actorRole) {
		f.ScopeUserID = actorID
	}

	// Apply plant scope: non-SuperAdmin users can only see issues from their own plant
	userPlantID, _ := c.Locals("userPlantID").(string)
	if userPlantID != "" {
		f.PlantID = userPlantID
	}
}

// ExportFiltered exports every issue matching the same filters and access scope as the list endpoint.
func (h *IssueFilterHandler) ExportFiltered(c *fiber.Ctx) error {
	f, err := issue.NewIssueFilter(c)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	applyIssueRequestScope(c, f)

	items, err := h.uc.GetForExport(f, issueExportLimit)
	if err != nil {
		return response.BadRequest(c, "gagal menyiapkan ekspor temuan", err.Error())
	}
	if len(items) == 0 {
		return response.BadRequest(c, "tidak ada data temuan yang sesuai dengan filter", nil)
	}

	statusCounts := make(map[string]int)
	payload := &exporter.PlaceholderPayload{
		Headers: map[string]interface{}{
			"generated_at":  time.Now().In(jakartaExportLocation()).Format("02 January 2006 15:04 WIB"),
			"period_filter": issueExportPeriod(c.Query("date_from"), c.Query("date_to")),
			"status_filter": exportText(c.Query("status"), "Semua status"),
			"search_filter": exportText(c.Query("q"), "Tanpa kata pencarian"),
			"total":         len(items),
		},
		Items: make([]map[string]interface{}, 0, len(items)),
	}

	for index, item := range items {
		initialCount, followUpCount, wowrCount := 0, 0, 0
		for _, photo := range item.Photos {
			switch photo.PhotoType {
			case issue.PhotoTypeInitial:
				initialCount++
			case issue.PhotoTypeFollowUp:
				followUpCount++
			case issue.PhotoTypeWOWR:
				wowrCount++
			}
		}

		status := item.ComputedIssueStatus
		if status == "" {
			status = item.IssueStatus
		}
		statusCounts[string(status)]++

		heiDetail := strings.TrimSpace(strings.Join(nonEmptyIssueExport(item.HEICategory, item.HEIName), " - "))
		payload.Items = append(payload.Items, map[string]interface{}{
			"no":                  index + 1,
			"issue_id":            item.IssueID,
			"created_at":          item.IssueCreatedAt.In(jakartaExportLocation()).Format("02-Jan-2006 15:04"),
			"plant":               exportText(item.PlantName, "-"),
			"area":                exportText(item.AreaName, "-"),
			"kawasan":             exportText(item.KawasanName, "-"),
			"detail_kawasan":      exportText(item.DetailKawasanName, "-"),
			"aspek":               exportText(item.AspekName, "-"),
			"detail_aspek":        exportText(item.DetailAspekName, "-"),
			"pic":                 exportText(item.PICName, item.IssuePICUserID),
			"status":              status,
			"due_date":            formatIssueExportDate(item.DueDate),
			"follow_up_delay":     exportDelay(item.FollowUpDelay),
			"needs_wowr":          map[bool]string{true: "Ya", false: "Tidak"}[item.NeedsWOWR],
			"wo_id":               exportText(item.WO_ID, "-"),
			"wr_id":               exportText(item.WR_ID, "-"),
			"wowr_status":         exportText(string(item.WOWRStatus), "None"),
			"hei_detail":          exportText(heiDetail, "-"),
			"finding_description": exportText(item.UraianText, item.Keterangan),
			"keterangan":          exportText(item.Keterangan, "-"),
			"initial_count":       initialCount,
			"follow_up_count":     followUpCount,
			"wowr_count":          wowrCount,
			"updated_at":          item.IssueUpdatedAt.In(jakartaExportLocation()).Format("02-Jan-2006 15:04"),
		})
	}

	payload.Headers["status_summary"] = issueStatusSummary(statusCounts)
	payload.TotalRow = &exporter.TableTotalRow{
		Label:            "TOTAL KESELURUHAN TEMUAN",
		LabelStartColumn: "A",
		LabelEndColumn:   "W",
		ValueColumn:      "X",
		Value:            len(items),
	}
	buf, err := exporter.GenerateExcelWithPlaceholder("./templates/issue_report.xlsx", "Laporan Temuan", payload)
	if err != nil {
		return response.InternalServerError(c, "gagal membuat file Excel temuan", err.Error())
	}

	fileName := fmt.Sprintf("Temuan_Inspeksi_%s.xlsx", time.Now().In(jakartaExportLocation()).Format("20060102_150405"))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	return c.SendStream(buf)
}

func jakartaExportLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return location
}

func issueExportPeriod(from, to string) string {
	if from == "" && to == "" {
		return "Semua periode"
	}
	return fmt.Sprintf("%s s.d. %s", exportText(from, "Awal"), exportText(to, "Sekarang"))
}

func exportText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func nonEmptyIssueExport(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

func formatIssueExportDate(value *time.Time) string {
	if value == nil {
		return "-"
	}
	return value.In(jakartaExportLocation()).Format("02-Jan-2006")
}

func exportDelay(value *int) string {
	if value == nil || *value <= 0 {
		return "0 hari"
	}
	return fmt.Sprintf("%d hari", *value)
}

func issueStatusSummary(counts map[string]int) string {
	order := []string{"Open", "InProgress", "PendingValidation", "OpenOverdue", "Closed", "Verified", "ClosedOverdue"}
	parts := make([]string, 0, len(order))
	for _, status := range order {
		if count := counts[status]; count > 0 {
			parts = append(parts, fmt.Sprintf("%s: %d", status, count))
		}
	}
	return strings.Join(parts, " | ")
}

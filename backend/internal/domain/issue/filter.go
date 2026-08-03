package issue

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/pagination"
	"gorm.io/gorm"
)

// ─── Shared Facet Types ───────────────────────────────────────────────────

// DateRangeFacet holds min/max for a date field.
type DateRangeFacet struct {
	Min   *time.Time `json:"min"`
	Max   *time.Time `json:"max"`
	Field string     `json:"field"`
}

// IssueFacets holds the facets returned alongside issue filter results.
type IssueFacets struct {
	Status       map[string]int64 `json:"status"`
	WOWRStatus   map[string]int64 `json:"wowr_status"`
	PICUserID    map[string]int64 `json:"issue_pic_user_id"`
	DateRange    DateRangeFacet   `json:"date_range"`
	DueDateRange DateRangeFacet   `json:"due_date_range"`
}

// ─── IssueFilter struct ───────────────────────────────────────────────────

// IssueFilter holds all filter parameters for the /issues/filter endpoint.
type IssueFilter struct {
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
	Q         string

	Status       string
	StatusIn     []string
	PICUserID    string
	PICUserIDIn  []string
	WOWRStatus   string
	WOWRStatusIn []string
	NeedsWOWR    *bool
	Label        string

	DateFrom *time.Time
	DateTo   *time.Time
	DueFrom  *time.Time
	DueTo    *time.Time

	// Scope override: if set, only issues with IssuePICUserID==ScopeUserID or delegated to ScopeUserID
	ScopeUserID string

	// PlantID scope: if set (non-SuperAdmin), only issues whose inspection area belongs to this plant
	PlantID string
}

var issueSortWhitelist = map[string]string{
	"created_at": `"IssueCreatedAt"`,
	"updated_at": `"IssueUpdatedAt"`,
	"due_date":   `"DueDate"`,
	"status":     `"IssueStatus"`,
}

func (f *IssueFilter) sortExpr() string {
	col, ok := issueSortWhitelist[f.SortBy]
	if !ok {
		col = `"IssueCreatedAt"`
	}
	order := "DESC"
	if strings.ToUpper(f.SortOrder) == "ASC" {
		order = "ASC"
	}
	return col + " " + order
}

func escapeLike(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)
	return "%" + strings.ToLower(q) + "%"
}

// ApplyTo builds a GORM query for the issue filter.
func (f *IssueFilter) ApplyTo(q *gorm.DB) *gorm.DB {
	// Plant scope enforcement: non-SuperAdmin users can only see issues from their plant's areas
	if f.PlantID != "" {
		q = q.Where(
			`"IssueID" IN (
				SELECT i."IssueID" FROM "Issue" i
				JOIN "Inspection_Result" ir ON i."ResultID" = ir."ResultID"
				JOIN "Inspection_Header" ih ON ir."InspectionID" = ih."InspectionID"
				JOIN "Area_Master" am ON ih."AreaID" = am."AreaID"
				WHERE am."PlantID" = ?
			)`,
			f.PlantID,
		)
	}

	// Scope enforcement for auditee
	if f.ScopeUserID != "" {
		q = q.Where(
			`"IssuePICUserID" = ? 
			OR "IssueID" IN (SELECT "IssueID" FROM "Issue_Delegate" WHERE "DelegateUserID" = ?)
			OR "IssueID" IN (
				SELECT i."IssueID" FROM "Issue" i
				JOIN "Inspection_Result" ir ON i."ResultID" = ir."ResultID"
				JOIN "Inspection_Header" ih ON ir."InspectionID" = ih."InspectionID"
				JOIN "PIC_Mapping" pm ON ih."KawasanID" = pm."KawasanID"
				WHERE pm."UserID" = ?
			)`,
			f.ScopeUserID, f.ScopeUserID, f.ScopeUserID,
		)
	}

	if len(f.StatusIn) > 0 {
		q = q.Where(`"IssueStatus" IN ?`, f.StatusIn)
	} else if f.Status != "" {
		q = q.Where(`"IssueStatus" = ?`, f.Status)
	}

	if len(f.PICUserIDIn) > 0 {
		q = q.Where(`"IssuePICUserID" IN ?`, f.PICUserIDIn)
	} else if f.PICUserID != "" {
		q = q.Where(`"IssuePICUserID" = ?`, f.PICUserID)
	}

	if len(f.WOWRStatusIn) > 0 {
		q = q.Where(`"WOWRStatus" IN ?`, f.WOWRStatusIn)
	} else if f.WOWRStatus != "" {
		q = q.Where(`"WOWRStatus" = ?`, f.WOWRStatus)
	}

	if f.NeedsWOWR != nil {
		q = q.Where(`"NeedsWOWR" = ?`, *f.NeedsWOWR)
	}

	if f.Label != "" {
		q = q.Where(`LOWER("Label") LIKE ?`, escapeLike(f.Label))
	}

	// Free-text search on IssueID, Label, WO_ID, WR_ID, AreaName, KawasanName, DetailKawasanName, PICName
	if f.Q != "" {
		likePattern := escapeLike(f.Q)
		q = q.Where(
			`LOWER("Issue"."IssueID") LIKE ? OR `+
			`LOWER(COALESCE("Issue"."Label", '')) LIKE ? OR `+
			`LOWER(COALESCE("Issue"."WO_ID", '')) LIKE ? OR `+
			`LOWER(COALESCE("Issue"."WR_ID", '')) LIKE ? OR `+
			`LOWER(COALESCE((SELECT u."FullName" FROM "Users" u WHERE u."UserID" = "Issue"."IssuePICUserID"), '')) LIKE ? OR `+
			`LOWER(COALESCE((
				SELECT am."AreaName" 
				FROM "Inspection_Result" ir
				JOIN "Inspection_Header" ih ON ir."InspectionID" = ih."InspectionID"
				JOIN "Area_Master" am ON ih."AreaID" = am."AreaID"
				WHERE ir."ResultID" = "Issue"."ResultID"
			), '')) LIKE ? OR `+
			`LOWER(COALESCE((
				SELECT km."KawasanName" 
				FROM "Inspection_Result" ir
				JOIN "Inspection_Header" ih ON ir."InspectionID" = ih."InspectionID"
				JOIN "Kawasan_Master" km ON ih."KawasanID" = km."KawasanID"
				WHERE ir."ResultID" = "Issue"."ResultID"
			), '')) LIKE ? OR `+
			`LOWER(COALESCE((
				SELECT dkm."DetailKawasanName" 
				FROM "Inspection_Result" ir
				JOIN "Inspection_Header" ih ON ir."InspectionID" = ih."InspectionID"
				JOIN "DetailKawasan_Master" dkm ON ih."DetailKawasanID" = dkm."DetailKawasanID"
				WHERE ir."ResultID" = "Issue"."ResultID"
			), '')) LIKE ?`,
			likePattern, likePattern, likePattern, likePattern, likePattern, likePattern, likePattern, likePattern,
		)
	}

	if f.DateFrom != nil {
		q = q.Where(`"IssueCreatedAt" >= ?`, f.DateFrom)
	}
	if f.DateTo != nil {
		q = q.Where(`"IssueCreatedAt" <= ?`, f.DateTo)
	}
	if f.DueFrom != nil {
		q = q.Where(`"DueDate" >= ?`, f.DueFrom)
	}
	if f.DueTo != nil {
		q = q.Where(`"DueDate" <= ?`, f.DueTo)
	}

	return q
}

// ApplySort applies the ORDER BY clause.
func (f *IssueFilter) ApplySort(q *gorm.DB) *gorm.DB {
	return q.Order(f.sortExpr())
}

// FiltersApplied echoes back applied params.
func (f *IssueFilter) FiltersApplied() map[string]interface{} {
	applied := map[string]interface{}{}
	if f.Status != "" {
		applied["status"] = []string{f.Status}
	}
	if len(f.StatusIn) > 0 {
		applied["status"] = f.StatusIn
	}
	if f.PICUserID != "" {
		applied["issue_pic_user_id"] = []string{f.PICUserID}
	}
	if len(f.PICUserIDIn) > 0 {
		applied["issue_pic_user_id"] = f.PICUserIDIn
	}
	if f.WOWRStatus != "" {
		applied["wowr_status"] = []string{f.WOWRStatus}
	}
	if f.Label != "" {
		applied["label"] = f.Label
	}
	if f.Q != "" {
		applied["q"] = f.Q
	}
	if f.DateFrom != nil {
		applied["date_from"] = f.DateFrom.Format(time.RFC3339)
	}
	if f.DateTo != nil {
		applied["date_to"] = f.DateTo.Format(time.RFC3339)
	}
	if f.DueFrom != nil {
		applied["due_from"] = f.DueFrom.Format(time.RFC3339)
	}
	if f.DueTo != nil {
		applied["due_to"] = f.DueTo.Format(time.RFC3339)
	}
	return applied
}

// ─── Repository & UseCase Interfaces ─────────────────────────────────────

// IssueFilterResult is the full paginated + faceted response.
type IssueFilterResult struct {
	Items          []Issue                `json:"items"`
	Total          int64                  `json:"total"`
	Page           int                    `json:"page"`
	Limit          int                    `json:"limit"`
	TotalPages     int64                  `json:"total_pages"`
	FiltersApplied map[string]interface{} `json:"filters_applied"`
	Facets         IssueFacets            `json:"facets"`
}

// IssueFilterRepository is implemented by issuerepo.
type IssueFilterRepository interface {
	FindFiltered(f *IssueFilter) ([]Issue, int64, error)
	FindFacets(f *IssueFilter) (IssueFacets, error)
}

// IssueFilterUseCase is implemented by issueusecase.
type IssueFilterUseCase interface {
	GetFiltered(f *IssueFilter) (*IssueFilterResult, error)
}

// ─── Query Helper ─────────────────────────────────────────────────────────

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDate(v string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", v)
}

// NewIssueFilter parses an IssueFilter from a Fiber request context.
func NewIssueFilter(c *fiber.Ctx) *IssueFilter {
	p := pagination.FromQuery(c)

	f := &IssueFilter{
		Page:       p.Page,
		Limit:      p.Limit,
		SortBy:     c.Query("sort_by", "created_at"),
		SortOrder:  c.Query("sort_order", "desc"),
		Q:          c.Query("q"),
		Status:     c.Query("status"),
		PICUserID:  c.Query("issue_pic_user_id"),
		WOWRStatus: c.Query("wowr_status"),
		Label:      c.Query("label"),
	}

	if v := c.Query("status__in"); v != "" {
		f.StatusIn = splitCSV(v)
	}
	if v := c.Query("issue_pic_user_id__in"); v != "" {
		f.PICUserIDIn = splitCSV(v)
	}
	if v := c.Query("wowr_status__in"); v != "" {
		f.WOWRStatusIn = splitCSV(v)
	}

	if nw := c.Query("needs_wo_wr"); nw != "" {
		b := nw == "true" || nw == "1"
		f.NeedsWOWR = &b
	}

	if v := c.Query("date_from"); v != "" {
		if t, err := parseDate(v); err == nil {
			f.DateFrom = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := parseDate(v); err == nil {
			f.DateTo = &t
		}
	}
	if v := c.Query("due_from"); v != "" {
		if t, err := parseDate(v); err == nil {
			f.DueFrom = &t
		}
	}
	if v := c.Query("due_to"); v != "" {
		if t, err := parseDate(v); err == nil {
			f.DueTo = &t
		}
	}

	return f
}

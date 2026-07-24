package issue

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/pagination"
	"gorm.io/gorm"
)

// FollowupFacets holds the facets returned for /issues/followup/filter.
type FollowupFacets struct {
	Status     map[string]int64 `json:"status"`
	WOWRStatus map[string]int64 `json:"wowr_status"`
	DateRange  DateRangeFacet   `json:"date_range"`
	DueDateRange DateRangeFacet `json:"due_date_range"`
}

// IssueFollowupItem enriches Issue with a computed IsDelegated field.
type IssueFollowupItem struct {
	Issue
	IsDelegated bool `json:"is_delegated" gorm:"-"`
}

// FollowupFilter holds filter parameters for /issues/followup/filter.
type FollowupFilter struct {
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
	Q         string

	Status       string
	StatusIn     []string
	WOWRStatus   string
	WOWRStatusIn []string
	NeedsWOWR    *bool
	Label        string

	IncludeClosed bool

	DateFrom *time.Time
	DateTo   *time.Time
	DueFrom  *time.Time
	DueTo    *time.Time

	// Scope: auditee → only their own (PIC or delegated)
	ScopeUserID string
}

var followupSortWhitelist = map[string]string{
	"created_at": `"Issue"."IssueCreatedAt"`,
	"updated_at": `"Issue"."IssueUpdatedAt"`,
	"due_date":   `"Issue"."DueDate"`,
	"status":     `"Issue"."IssueStatus"`,
}

func (f *FollowupFilter) sortExpr() string {
	col, ok := followupSortWhitelist[f.SortBy]
	if !ok {
		col = `"Issue"."IssueCreatedAt"`
	}
	order := "DESC"
	if strings.ToUpper(f.SortOrder) == "ASC" {
		order = "ASC"
	}
	return col + " " + order
}

// ApplyTo applies scope and filter conditions to a GORM query.
// Assumes the base query is on model Issue with optional join to Issue_Delegate.
func (f *FollowupFilter) ApplyTo(q *gorm.DB) *gorm.DB {
	// Default: exclude terminal statuses unless IncludeClosed
	if !f.IncludeClosed {
		q = q.Where(`"Issue"."IssueStatus" NOT IN ?`, []string{"Verified", "Closed"})
	}

	// Scope enforcement for auditee
	if f.ScopeUserID != "" {
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
			f.ScopeUserID, f.ScopeUserID, f.ScopeUserID,
		)
	}

	if len(f.StatusIn) > 0 {
		q = q.Where(`"Issue"."IssueStatus" IN ?`, f.StatusIn)
	} else if f.Status != "" {
		q = q.Where(`"Issue"."IssueStatus" = ?`, f.Status)
	}

	if len(f.WOWRStatusIn) > 0 {
		q = q.Where(`"Issue"."WOWRStatus" IN ?`, f.WOWRStatusIn)
	} else if f.WOWRStatus != "" {
		q = q.Where(`"Issue"."WOWRStatus" = ?`, f.WOWRStatus)
	}

	if f.NeedsWOWR != nil {
		q = q.Where(`"Issue"."NeedsWOWR" = ?`, *f.NeedsWOWR)
	}

	if f.Label != "" {
		q = q.Where(`LOWER("Issue"."Label") LIKE ?`, escapeLike(f.Label))
	}
	if f.Q != "" {
		q = q.Where(`LOWER("Issue"."Label") LIKE ?`, escapeLike(f.Q))
	}

	if f.DateFrom != nil {
		q = q.Where(`"Issue"."IssueCreatedAt" >= ?`, f.DateFrom)
	}
	if f.DateTo != nil {
		q = q.Where(`"Issue"."IssueCreatedAt" <= ?`, f.DateTo)
	}
	if f.DueFrom != nil {
		q = q.Where(`"Issue"."DueDate" >= ?`, f.DueFrom)
	}
	if f.DueTo != nil {
		q = q.Where(`"Issue"."DueDate" <= ?`, f.DueTo)
	}

	return q
}

// ApplySort applies ORDER BY.
func (f *FollowupFilter) ApplySort(q *gorm.DB) *gorm.DB {
	return q.Order(f.sortExpr())
}

// FiltersApplied echoes back the applied params.
func (f *FollowupFilter) FiltersApplied() map[string]interface{} {
	applied := map[string]interface{}{}
	if f.Status != "" {
		applied["status"] = []string{f.Status}
	}
	if len(f.StatusIn) > 0 {
		applied["status"] = f.StatusIn
	}
	if f.WOWRStatus != "" {
		applied["wowr_status"] = []string{f.WOWRStatus}
	}
	if f.IncludeClosed {
		applied["include_closed"] = true
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

// FollowupFilterResult is the full paginated + faceted response for followup.
type FollowupFilterResult struct {
	Items          []Issue                `json:"items"`
	Total          int64                  `json:"total"`
	Page           int                    `json:"page"`
	Limit          int                    `json:"limit"`
	TotalPages     int64                  `json:"total_pages"`
	FiltersApplied map[string]interface{} `json:"filters_applied"`
	Facets         FollowupFacets         `json:"facets"`
}

// FollowupFilterRepository is implemented by issuerepo.
type FollowupFilterRepository interface {
	FindFiltered(f *FollowupFilter) ([]Issue, int64, error)
	FindFacets(f *FollowupFilter) (FollowupFacets, error)
}

// FollowupFilterUseCase is implemented by issueusecase.
type FollowupFilterUseCase interface {
	GetFiltered(f *FollowupFilter) (*FollowupFilterResult, error)
}

// ─── Query Helper ─────────────────────────────────────────────────────────

// NewFollowupFilter parses a FollowupFilter from a Fiber context.
func NewFollowupFilter(c *fiber.Ctx) *FollowupFilter {
	p := pagination.FromQuery(c)

	f := &FollowupFilter{
		Page:          p.Page,
		Limit:         p.Limit,
		SortBy:        c.Query("sort_by", "created_at"),
		SortOrder:     c.Query("sort_order", "desc"),
		Q:             c.Query("q"),
		Status:        c.Query("status"),
		WOWRStatus:    c.Query("wowr_status"),
		Label:         c.Query("label"),
		IncludeClosed: c.Query("include_closed") == "true" || c.Query("include_closed") == "1",
	}

	if v := c.Query("status__in"); v != "" {
		f.StatusIn = splitCSV(v)
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

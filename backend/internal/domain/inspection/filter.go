package inspection

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/pagination"
	"gorm.io/gorm"
)

// ─── Shared Facet Types ───────────────────────────────────────────────────

// DateRangeFacet holds the min/max value for a date field.
type DateRangeFacet struct {
	Min   *time.Time `json:"min"`
	Max   *time.Time `json:"max"`
	Field string     `json:"field"`
}

// FacetMap is field → value → count.
type FacetMap map[string]map[string]int64

// InspectionFacets holds the facets returned alongside inspection filter results.
type InspectionFacets struct {
	Status     map[string]int64 `json:"status"`
	AreaID     map[string]int64 `json:"area_id"`
	KawasanID  map[string]int64 `json:"kawasan_id"`
	InspectorID map[string]int64 `json:"inspector_id"`
	DateRange  DateRangeFacet   `json:"date_range"`
}

// ─── InspectionFilter struct ─────────────────────────────────────────────

// InspectionFilter holds all filter parameters for the /inspections/filter endpoint.
type InspectionFilter struct {
	// Pagination
	Page  int
	Limit int

	// Sort
	SortBy    string
	SortOrder string

	// Free-text
	Q string

	// Exact / multi-value
	Status          string
	StatusIn        []string
	AreaID          string
	AreaIDIn        []string
	KawasanID       string
	KawasanIDIn     []string
	DetailKawasanID string
	InspectorID     string
	InspectorIDIn   []string

	// Date range (created_at)
	DateFrom *time.Time
	DateTo   *time.Time

	// Scope override (injected by handler for auditee)
	ScopeInspectorID string
}

var inspectionSortWhitelist = map[string]string{
	"created_at": `"InspectionHeaderCreatedAt"`,
	"updated_at": `"InspectionheaderUpdatedAt"`,
	"status":     `"InspectionHeaderStatus"`,
}

// AllowedSortFields returns the whitelisted sort fields.
func (f *InspectionFilter) AllowedSortFields() []string {
	keys := make([]string, 0, len(inspectionSortWhitelist))
	for k := range inspectionSortWhitelist {
		keys = append(keys, k)
	}
	return keys
}

// sortExpr returns the safe ORDER BY expression.
func (f *InspectionFilter) sortExpr() string {
	col, ok := inspectionSortWhitelist[f.SortBy]
	if !ok {
		col = `"InspectionHeaderCreatedAt"`
	}
	order := "DESC"
	if strings.ToUpper(f.SortOrder) == "ASC" {
		order = "ASC"
	}
	return col + " " + order
}

// escapeLike escapes special LIKE characters in q to avoid injection.
func escapeLike(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)
	return "%" + strings.ToLower(q) + "%"
}

// ApplyTo builds a GORM query for the inspection filter (without pagination).
// Call this method and then apply Offset/Limit separately.
func (f *InspectionFilter) ApplyTo(q *gorm.DB) *gorm.DB {
	// Scope override: auditee can only see their own inspections
	if f.ScopeInspectorID != "" {
		q = q.Where(`"InspectorID" = ?`, f.ScopeInspectorID)
	}

	// Exact match filters
	if len(f.StatusIn) > 0 {
		q = q.Where(`"InspectionHeaderStatus" IN ?`, f.StatusIn)
	} else if f.Status != "" {
		q = q.Where(`"InspectionHeaderStatus" = ?`, f.Status)
	}

	if len(f.AreaIDIn) > 0 {
		q = q.Where(`"AreaID" IN ?`, f.AreaIDIn)
	} else if f.AreaID != "" {
		q = q.Where(`"AreaID" = ?`, f.AreaID)
	}

	if len(f.KawasanIDIn) > 0 {
		q = q.Where(`"KawasanID" IN ?`, f.KawasanIDIn)
	} else if f.KawasanID != "" {
		q = q.Where(`"KawasanID" = ?`, f.KawasanID)
	}

	if f.DetailKawasanID != "" {
		q = q.Where(`"DetailKawasanID" = ?`, f.DetailKawasanID)
	}

	if len(f.InspectorIDIn) > 0 {
		q = q.Where(`"InspectorID" IN ?`, f.InspectorIDIn)
	} else if f.InspectorID != "" {
		q = q.Where(`"InspectorID" = ?`, f.InspectorID)
	}

	// Date range
	if f.DateFrom != nil {
		q = q.Where(`"InspectionHeaderCreatedAt" >= ?`, f.DateFrom)
	}
	if f.DateTo != nil {
		q = q.Where(`"InspectionHeaderCreatedAt" <= ?`, f.DateTo)
	}

	// Free-text search on InspectionID, SessionID, AreaName, KawasanName, DetailKawasanName
	if f.Q != "" {
		likePattern := escapeLike(f.Q)
		q = q.Where(
			`LOWER("Inspection_Header"."InspectionID") LIKE ? OR `+
			`LOWER(COALESCE("Inspection_Header"."SessionID", '')) LIKE ? OR `+
			`LOWER(COALESCE((SELECT "AreaName" FROM "Area_Master" WHERE "Area_Master"."AreaID" = "Inspection_Header"."AreaID"), '')) LIKE ? OR `+
			`LOWER(COALESCE((SELECT "KawasanName" FROM "Kawasan_Master" WHERE "Kawasan_Master"."KawasanID" = "Inspection_Header"."KawasanID"), '')) LIKE ? OR `+
			`LOWER(COALESCE((SELECT "DetailKawasanName" FROM "DetailKawasan_Master" WHERE "DetailKawasan_Master"."DetailKawasanID" = "Inspection_Header"."DetailKawasanID"), '')) LIKE ?`,
			likePattern, likePattern, likePattern, likePattern, likePattern,
		)
	}

	return q
}

// ApplySort applies the ORDER BY clause.
func (f *InspectionFilter) ApplySort(q *gorm.DB) *gorm.DB {
	return q.Order(f.sortExpr())
}

// FiltersApplied returns the map of applied filters for debugging in the response.
func (f *InspectionFilter) FiltersApplied() map[string]interface{} {
	applied := map[string]interface{}{}
	if f.Status != "" {
		applied["status"] = []string{f.Status}
	}
	if len(f.StatusIn) > 0 {
		applied["status"] = f.StatusIn
	}
	if f.AreaID != "" {
		applied["area_id"] = []string{f.AreaID}
	}
	if len(f.AreaIDIn) > 0 {
		applied["area_id"] = f.AreaIDIn
	}
	if f.KawasanID != "" {
		applied["kawasan_id"] = []string{f.KawasanID}
	}
	if len(f.KawasanIDIn) > 0 {
		applied["kawasan_id"] = f.KawasanIDIn
	}
	if f.InspectorID != "" {
		applied["inspector_id"] = []string{f.InspectorID}
	}
	if len(f.InspectorIDIn) > 0 {
		applied["inspector_id"] = f.InspectorIDIn
	}
	if f.DetailKawasanID != "" {
		applied["detail_kawasan_id"] = []string{f.DetailKawasanID}
	}
	if f.DateFrom != nil {
		applied["date_from"] = f.DateFrom.Format(time.RFC3339)
	}
	if f.DateTo != nil {
		applied["date_to"] = f.DateTo.Format(time.RFC3339)
	}
	if f.Q != "" {
		applied["q"] = f.Q
	}
	return applied
}

// ─── Repository & UseCase Interfaces ─────────────────────────────────────

// InspectionFilterResult is the full paginated + faceted response.
type InspectionFilterResult struct {
	Items          []InspectionHeader     `json:"items"`
	Total          int64                  `json:"total"`
	Page           int                    `json:"page"`
	Limit          int                    `json:"limit"`
	TotalPages     int64                  `json:"total_pages"`
	FiltersApplied map[string]interface{} `json:"filters_applied"`
	Facets         InspectionFacets       `json:"facets"`
}

// InspectionFilterRepository is implemented by inspectionrepo.
type InspectionFilterRepository interface {
	FindFiltered(f *InspectionFilter) ([]InspectionHeader, int64, error)
	FindFacets(f *InspectionFilter) (InspectionFacets, error)
}

// InspectionFilterUseCase is implemented by inspectionusecase.
type InspectionFilterUseCase interface {
	GetFiltered(f *InspectionFilter) (*InspectionFilterResult, error)
}

// ─── Query Helper ─────────────────────────────────────────────────────────

// NewInspectionFilter parses an InspectionFilter from a Fiber request context.
func NewInspectionFilter(c *fiber.Ctx) *InspectionFilter {
	p := pagination.FromQuery(c)

	f := &InspectionFilter{
		Page:            p.Page,
		Limit:           p.Limit,
		SortBy:          c.Query("sort_by", "created_at"),
		SortOrder:       c.Query("sort_order", "desc"),
		Q:               c.Query("q"),
		Status:          c.Query("status"),
		AreaID:          c.Query("area_id"),
		KawasanID:       c.Query("kawasan_id"),
		DetailKawasanID: c.Query("detail_kawasan_id"),
		InspectorID:     c.Query("inspector_id"),
	}

	if v := c.Query("status__in"); v != "" {
		f.StatusIn = splitCSV(v)
	}
	if v := c.Query("area_id__in"); v != "" {
		f.AreaIDIn = splitCSV(v)
	}
	if v := c.Query("kawasan_id__in"); v != "" {
		f.KawasanIDIn = splitCSV(v)
	}
	if v := c.Query("inspector_id__in"); v != "" {
		f.InspectorIDIn = splitCSV(v)
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

	return f
}

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
	// Try RFC3339 first, then date-only format
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", v)
}

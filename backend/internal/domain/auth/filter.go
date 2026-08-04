package auth

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/pagination"
	"gorm.io/gorm"
)

// UserFacets holds the facets returned alongside user filter results.
type UserFacets struct {
	RoleID       map[string]int64 `json:"role_id"`
	DepartmentID map[string]int64 `json:"department_id"`
	PlantID      map[string]int64 `json:"plant_id"`
	UserStatus   map[string]int64 `json:"user_status"`
	DateRange    struct {
		Min   *time.Time `json:"min"`
		Max   *time.Time `json:"max"`
		Field string     `json:"field"`
	} `json:"date_range"`
}

// UserFilter holds all filter parameters for the /users/filter endpoint.
type UserFilter struct {
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
	Q         string

	RoleID       string
	RoleIDIn     []string
	DepartmentID string
	PlantID      string
	PlantIDIn    []string
	UserStatus   string
	UserStatusIn []string

	DateFrom *time.Time
	DateTo   *time.Time
}

var userSortWhitelist = map[string]string{
	"created_at": `"UserCreatedAt"`,
	"updated_at": `"UserUpdatedAt"`,
	"username":   `"Username"`,
	"full_name":  `"FullName"`,
}

func (f *UserFilter) sortExpr() string {
	col, ok := userSortWhitelist[f.SortBy]
	if !ok {
		col = `"UserCreatedAt"`
	}
	order := "DESC"
	if strings.ToUpper(f.SortOrder) == "ASC" {
		order = "ASC"
	}
	return col + " " + order
}

func userEscapeLike(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)
	return "%" + strings.ToLower(q) + "%"
}

// ApplyTo builds a GORM query for the user filter.
func (f *UserFilter) ApplyTo(q *gorm.DB) *gorm.DB {
	if f.Q != "" {
		like := userEscapeLike(f.Q)
		q = q.Where(`LOWER("Username") LIKE ? OR LOWER("FullName") LIKE ? OR LOWER("Email") LIKE ?`, like, like, like)
	}

	if len(f.RoleIDIn) > 0 {
		q = q.Where(`"RoleID" IN ?`, f.RoleIDIn)
	} else if f.RoleID != "" {
		q = q.Where(`"RoleID" = ?`, f.RoleID)
	}

	if f.DepartmentID != "" {
		q = q.Where(`"DepartmentID" = ?`, f.DepartmentID)
	}

	if len(f.PlantIDIn) > 0 {
		q = q.Where(`"PlantID" IN ?`, f.PlantIDIn)
	} else if f.PlantID != "" {
		if f.PlantID == "GLOBAL" || f.PlantID == "NULL" {
			q = q.Where(`"PlantID" IS NULL OR "PlantID" = ''`)
		} else {
			q = q.Where(`"PlantID" = ?`, f.PlantID)
		}
	}

	if len(f.UserStatusIn) > 0 {
		q = q.Where(`"UserStatus" IN ?`, f.UserStatusIn)
	} else if f.UserStatus != "" {
		q = q.Where(`"UserStatus" = ?`, f.UserStatus)
	}

	if f.DateFrom != nil {
		q = q.Where(`"UserCreatedAt" >= ?`, f.DateFrom)
	}
	if f.DateTo != nil {
		q = q.Where(`"UserCreatedAt" <= ?`, f.DateTo)
	}

	return q
}

// ApplySort applies the ORDER BY clause.
func (f *UserFilter) ApplySort(q *gorm.DB) *gorm.DB {
	return q.Order(f.sortExpr())
}

// FiltersApplied echoes back applied params.
func (f *UserFilter) FiltersApplied() map[string]interface{} {
	applied := map[string]interface{}{}
	if f.Q != "" {
		applied["q"] = f.Q
	}
	if f.RoleID != "" {
		applied["role_id"] = []string{f.RoleID}
	}
	if len(f.RoleIDIn) > 0 {
		applied["role_id"] = f.RoleIDIn
	}
	if f.DepartmentID != "" {
		applied["department_id"] = []string{f.DepartmentID}
	}
	if f.PlantID != "" {
		applied["plant_id"] = []string{f.PlantID}
	}
	if len(f.PlantIDIn) > 0 {
		applied["plant_id"] = f.PlantIDIn
	}
	if f.UserStatus != "" {
		applied["user_status"] = []string{f.UserStatus}
	}
	if len(f.UserStatusIn) > 0 {
		applied["user_status"] = f.UserStatusIn
	}
	if f.DateFrom != nil {
		applied["date_from"] = f.DateFrom.Format(time.RFC3339)
	}
	if f.DateTo != nil {
		applied["date_to"] = f.DateTo.Format(time.RFC3339)
	}
	return applied
}

// ─── Repository & UseCase Interfaces ─────────────────────────────────────

// UserFilterResult is the full paginated + faceted response.
type UserFilterResult struct {
	Items          []User                 `json:"items"`
	Total          int64                  `json:"total"`
	Page           int                    `json:"page"`
	Limit          int                    `json:"limit"`
	TotalPages     int64                  `json:"total_pages"`
	FiltersApplied map[string]interface{} `json:"filters_applied"`
	Facets         UserFacets             `json:"facets"`
}

// UserFilterRepository is implemented by authrepo.
type UserFilterRepository interface {
	FindFiltered(f *UserFilter) ([]User, int64, error)
	FindFacets(f *UserFilter) (UserFacets, error)
}

// UserFilterUseCase is implemented by authusecase.
type UserFilterUseCase interface {
	GetFiltered(f *UserFilter) (*UserFilterResult, error)
}

// ─── Query Helper ─────────────────────────────────────────────────────────

func userSplitCSV(v string) []string {
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

func userParseDate(v string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", v)
}

// NewUserFilter parses a UserFilter from a Fiber request context.
func NewUserFilter(c *fiber.Ctx) *UserFilter {
	p := pagination.FromQuery(c)

	f := &UserFilter{
		Page:         p.Page,
		Limit:        p.Limit,
		SortBy:       c.Query("sort_by", "created_at"),
		SortOrder:    c.Query("sort_order", "desc"),
		Q:            c.Query("q"),
		RoleID:       c.Query("role_id"),
		DepartmentID: c.Query("department_id"),
		PlantID:      c.Query("plant_id"),
		UserStatus:   c.Query("user_status"),
	}

	userPlantID, _ := c.Locals("userPlantID").(string)
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	if !isSuperAdmin && userPlantID != "" {
		f.PlantID = userPlantID
	}

	if v := c.Query("role_id__in"); v != "" {
		f.RoleIDIn = userSplitCSV(v)
	}
	if v := c.Query("plant_id__in"); v != "" {
		f.PlantIDIn = userSplitCSV(v)
	}
	if v := c.Query("user_status__in"); v != "" {
		f.UserStatusIn = userSplitCSV(v)
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := userParseDate(v); err == nil {
			f.DateFrom = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := userParseDate(v); err == nil {
			f.DateTo = &t
		}
	}

	return f
}

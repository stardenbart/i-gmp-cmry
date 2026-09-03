package dashboard

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// WidgetConfig describes one widget's placement on a user's dashboard.
// Order/Visible only for now (Fase 3a: checklist + numeric order, no drag).
// X/Y/W/H are reserved for Fase 3b (react-grid-layout) and are simply left
// zero-valued/omitted until then — existing rows stay valid once that lands.
type WidgetConfig struct {
	WidgetID string `json:"widget_id"`
	Visible  bool   `json:"visible"`
	Order    int    `json:"order"`

	// Reserved for the future drag/resize grid (Fase 3b). Omitted from JSON
	// (and thus absent from rows saved today) until the frontend starts
	// sending them; the grid falls back to Order-based flow layout when unset.
	X *int `json:"x,omitempty"`
	Y *int `json:"y,omitempty"`
	W *int `json:"w,omitempty"`
	H *int `json:"h,omitempty"`

	// VizType is which visualization the widget renders as (e.g. "table",
	// "bar" — see frontend components/dashboard/types.ts VizType). Opaque
	// to the backend: just round-tripped through LayoutJSON like X/Y/W/H,
	// no validation or schema change needed since this whole struct is
	// stored as a single JSON blob per row, not individual DB columns.
	VizType *string `json:"viz_type,omitempty"`

	// CustomQuery marks this widget as a user-built visualization from the
	// Custom KPI Visualization Builder (/kpi's "Tambah Visualisasi" panel).
	// Only ever present on
	// dashboard_key="kpi" rows. Unlike VizType, this IS validated — see
	// ValidateCustomQuery, called from SaveLayout — because it carries
	// user-chosen measure/dimension IDs that get sent straight to
	// POST /analytics/query at render time.
	CustomQuery *CustomQueryConfig `json:"custom_query,omitempty"`
}

// CustomQueryConfig is the persisted shape of one custom widget's semantic
// query selection. Deliberately carries only catalog IDs — never SQL,
// never a formula — so a saved layout can't smuggle anything the
// /analytics/query semantic validation wouldn't also re-check at query time.
type CustomQueryConfig struct {
	// Version guards future breaking changes to this shape; DynamicKPIWidget
	// must treat an unrecognized/future Version as "unavailable", not crash.
	Version   int      `json:"version"`
	Measures  []string `json:"measures"`
	Dimension string   `json:"dimension"`
	// Title gives a custom visualization its user-defined business meaning.
	// It is display-only and is never interpolated into an analytics query.
	// Empty remains accepted for layouts created before custom titles existed.
	Title string `json:"title,omitempty"`
	// Dimension2 is optional — only set for Heatmap/Sankey widgets (a
	// dim1 x dim2 x 1 measure matrix). "" for every other chart type.
	Dimension2 string `json:"dimension2,omitempty"`
	// DrillDimensions is optional — an ORDERED list of additional dimension
	// ids for drill-down (Dimension itself is level 0; DrillDimensions[0] is
	// level 1, etc). Mutually exclusive with Dimension2 (matrix mode and
	// drill-down are two different uses of "a second category" that don't
	// compose). Empty/nil for every widget built before this feature.
	DrillDimensions []string `json:"drill_dimensions,omitempty"`
	// ShowLabels/ShowLabelValues/ShowTrendLine are pure rendering options —
	// never sent to /analytics/query, never affect what rows come back, only
	// how DynamicKPIWidget draws them (see buildEChartsOption.ts). Kept here
	// anyway (not a separate WidgetConfig field) so a widget's whole visual
	// definition lives in one persisted bag, same reasoning as Dimension2/
	// DrillDimensions. Omitted (false) for every widget built before this
	// feature — see buildEChartsOption.ts's backward-compatible defaulting
	// for what "false" vs "absent" means per chart type.
	ShowLabels      bool `json:"show_labels,omitempty"`
	ShowLabelValues bool `json:"show_label_values,omitempty"`
	ShowTrendLine   bool `json:"show_trend_line,omitempty"`
}

// UserDashboardLayout is one user's saved dashboard configuration. One row
// per (user, dashboard) — dashboard layout is a personal preference, not
// plant-scoped like System_Setting. DashboardKey distinguishes a user's
// several independently-customizable dashboards (their role dashboard,
// "main"; the KPI/analytics dashboard, "kpi"; etc.) so saving one never
// clobbers another — see migration 043.
type UserDashboardLayout struct {
	UserID       string    `gorm:"column:UserID;primaryKey" json:"user_id"`
	DashboardKey string    `gorm:"column:DashboardKey;primaryKey" json:"dashboard_key"`
	LayoutJSON   string    `gorm:"column:LayoutJSON;type:text;not null" json:"-"`
	UpdatedAt    time.Time `gorm:"column:UpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (UserDashboardLayout) TableName() string { return "User_Dashboard_Layout" }

// maxCustomQueryMeasures mirrors analyticsusecase.MaxMeasuresPerQuery. Kept
// as its own small constant (rather than importing the usecase package
// here) to avoid a domain->usecase dependency — this package must stay
// pure per the layering rule. Change both together if the guardrail moves.
const maxCustomQueryMeasures = 4

// maxCustomQueryMeasuresTable is the higher ceiling that applies only when
// the widget's VizType is "table" — every other chart type stays at
// maxCustomQueryMeasures (more than 4 series is unreadable on a chart, but
// a table just gets another column). See ValidateCustomQuery.
const maxCustomQueryMeasuresTable = 30

// maxCustomQueryDrillLevels caps the extra drill-down levels beyond the
// primary Dimension (level 0), keeping the total (4) consistent with
// maxCustomQueryMeasures.
const maxCustomQueryDrillLevels = 3

const maxCustomQueryTitleLength = 100

var customQueryIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// ValidateCustomQuery enforces the same shape/size guardrails SaveLayout
// must apply BEFORE persisting a widget's custom_query — not just when
// /analytics/query runs it — so a saved layout can never hold
// oversized/garbage configuration. Field-existence in the live catalog is
// deliberately NOT checked here (that's analyticsusecase's job, at query
// time): a widget referencing a since-removed field must still be able to
// load its layout (and show "Konfigurasi widget sudah tidak tersedia"),
// not fail to save/reload entirely.
//
// vizType is the sibling WidgetConfig.VizType (nil for legacy rows saved
// before VizType existed, treated the same as "not table") — the only
// input here that lets this function tell Table apart from every other
// chart type, since CustomQueryConfig itself carries no chart-type info.
func ValidateCustomQuery(cq *CustomQueryConfig, vizType *string) error {
	if cq == nil {
		return nil
	}
	if cq.Version != 1 {
		return fmt.Errorf("custom_query version tidak didukung: %d", cq.Version)
	}
	cq.Title = strings.TrimSpace(cq.Title)
	if utf8.RuneCountInString(cq.Title) > maxCustomQueryTitleLength {
		return fmt.Errorf("custom_query title maksimal %d karakter", maxCustomQueryTitleLength)
	}
	if strings.IndexFunc(cq.Title, unicode.IsControl) >= 0 {
		return fmt.Errorf("custom_query title mengandung karakter kontrol")
	}
	if len(cq.Measures) == 0 {
		return fmt.Errorf("custom_query harus punya minimal 1 measure")
	}
	measuresLimit := maxCustomQueryMeasures
	if vizType != nil && *vizType == "table" {
		measuresLimit = maxCustomQueryMeasuresTable
	}
	if len(cq.Measures) > measuresLimit {
		return fmt.Errorf("custom_query maksimal %d measure", measuresLimit)
	}
	if !customQueryIDPattern.MatchString(cq.Dimension) {
		return fmt.Errorf("custom_query dimension tidak valid")
	}
	if cq.Dimension2 != "" && !customQueryIDPattern.MatchString(cq.Dimension2) {
		return fmt.Errorf("custom_query dimension2 tidak valid")
	}
	if len(cq.DrillDimensions) > 0 {
		if cq.Dimension2 != "" {
			return fmt.Errorf("custom_query drill_dimensions tidak bisa dipakai bersama dimension2")
		}
		if len(cq.DrillDimensions) > maxCustomQueryDrillLevels {
			return fmt.Errorf("custom_query maksimal %d level drill-down tambahan", maxCustomQueryDrillLevels)
		}
		seen := map[string]bool{cq.Dimension: true}
		for _, d := range cq.DrillDimensions {
			if !customQueryIDPattern.MatchString(d) {
				return fmt.Errorf("custom_query drill_dimensions tidak valid: %s", d)
			}
			if seen[d] {
				return fmt.Errorf("custom_query drill_dimensions punya duplikat: %s", d)
			}
			seen[d] = true
		}
	}
	for _, m := range cq.Measures {
		if !customQueryIDPattern.MatchString(m) {
			return fmt.Errorf("custom_query measure tidak valid: %s", m)
		}
	}
	return nil
}

// DashboardKeyMain is the implicit dashboard key every pre-existing layout
// row was migrated to (see migration 043) and the default whenever a
// caller doesn't specify ?dashboard_key= — this keeps the 3 existing role
// dashboards (and the "Edit User" dashboard tab) behaving exactly as
// before this multi-dashboard support was added.
const DashboardKeyMain = "main"

// DashboardKeyKPI is the dedicated KPI/analytics dashboard (Fase KPI).
const DashboardKeyKPI = "kpi"

// DashboardLayoutRepository persists the raw layout JSON blob per
// (user, dashboard).
type DashboardLayoutRepository interface {
	FindByUserID(userID, dashboardKey string) (*UserDashboardLayout, error)
	Upsert(userID, dashboardKey, layoutJSON string) error
}

// ── Well-known widget IDs ────────────────────────────────────────────────
// Fase 2 (frontend widget registry) is the source of truth for what each ID
// renders; these constants just need to stay in sync with it.
const (
	WidgetStatsCards        = "stats-cards"
	WidgetAktivitasInspeksi = "aktivitas-inspeksi"
	WidgetAktivitasPIC      = "aktivitas-pic"
	WidgetStatusWOWR        = "status-wowr"
	WidgetStatusAuditee     = "status-auditee"
	WidgetChartTrend        = "chart-tren-kepatuhan"

	WidgetAuditorStats         = "auditor-stats"
	WidgetAuditorPendingBanner = "auditor-pending-banner"
	WidgetAuditorTrendChart    = "auditor-trend-chart"

	WidgetAuditeeStats    = "auditee-stats"
	WidgetAuditeeTaskList = "auditee-task-list"
)

// DefaultLayoutForRole returns the out-of-the-box widget arrangement for a
// role, used whenever a user hasn't saved a customization yet. Keep this in
// sync with each role's current dashboard component during the Fase 2
// refactor — it's meant to reproduce today's fixed layout exactly, not
// introduce new defaults.
func DefaultLayoutForRole(roleID string) []WidgetConfig {
	switch roleID {
	case "ROLE-002": // Auditor
		return []WidgetConfig{
			{WidgetID: WidgetAuditorStats, Visible: true, Order: 0},
			{WidgetID: WidgetAuditorPendingBanner, Visible: true, Order: 1},
			{WidgetID: WidgetAuditorTrendChart, Visible: true, Order: 2},
		}
	case "ROLE-003": // Auditee
		return []WidgetConfig{
			{WidgetID: WidgetAuditeeStats, Visible: true, Order: 0},
			{WidgetID: WidgetAuditeeTaskList, Visible: true, Order: 1},
		}
	default: // ROLE-000 (Super Admin), ROLE-001 (Admin), and any other role
		return []WidgetConfig{
			{WidgetID: WidgetStatsCards, Visible: true, Order: 0},
			{WidgetID: WidgetAktivitasInspeksi, Visible: true, Order: 1},
			{WidgetID: WidgetAktivitasPIC, Visible: true, Order: 2},
			{WidgetID: WidgetStatusWOWR, Visible: true, Order: 3},
			{WidgetID: WidgetStatusAuditee, Visible: true, Order: 4},
			{WidgetID: WidgetChartTrend, Visible: true, Order: 5},
		}
	}
}

// DefaultLayoutKPI intentionally starts empty. The KPI dashboard is a
// workspace for user-created visualizations; fixed operational summaries
// remain on the main role dashboard and are not duplicated here.
func DefaultLayoutKPI() []WidgetConfig {
	return []WidgetConfig{}
}

// DefaultLayoutForDashboard resolves the out-of-the-box widget arrangement
// for a (dashboardKey, roleID) pair. Only DashboardKeyMain is role-specific
// (the 3 existing role dashboards); every other dashboard key has its own
// fixed default regardless of role — for now just DashboardKeyKPI, but this
// is the extension point for any future dedicated dashboard.
func DefaultLayoutForDashboard(dashboardKey, roleID string) []WidgetConfig {
	if dashboardKey == DashboardKeyKPI {
		return DefaultLayoutKPI()
	}
	return DefaultLayoutForRole(roleID)
}

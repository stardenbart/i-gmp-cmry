package dashboard

import "time"

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
}

// UserDashboardLayout is one user's saved dashboard configuration. One row
// per user — dashboard layout is a personal preference, not plant-scoped
// like System_Setting.
type UserDashboardLayout struct {
	UserID     string    `gorm:"column:UserID;primaryKey" json:"user_id"`
	LayoutJSON string    `gorm:"column:LayoutJSON;type:text;not null" json:"-"`
	UpdatedAt  time.Time `gorm:"column:UpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (UserDashboardLayout) TableName() string { return "User_Dashboard_Layout" }

// DashboardLayoutRepository persists the raw layout JSON blob per user.
type DashboardLayoutRepository interface {
	FindByUserID(userID string) (*UserDashboardLayout, error)
	Upsert(userID string, layoutJSON string) error
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

	WidgetAuditorInspeksiList = "auditor-inspeksi-list"
	WidgetAuditorTrendChart   = "auditor-trend-chart"
	WidgetAuditorIssueSummary = "auditor-issue-summary"
	WidgetAuditorPendingList  = "auditor-pending-list"

	WidgetAuditeeIssueList     = "auditee-issue-list"
	WidgetAuditeeComplianceSum = "auditee-compliance-summary"
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
			{WidgetID: WidgetAuditorInspeksiList, Visible: true, Order: 0},
			{WidgetID: WidgetAuditorTrendChart, Visible: true, Order: 1},
			{WidgetID: WidgetAuditorIssueSummary, Visible: true, Order: 2},
			{WidgetID: WidgetAuditorPendingList, Visible: true, Order: 3},
		}
	case "ROLE-003": // Auditee
		return []WidgetConfig{
			{WidgetID: WidgetAuditeeIssueList, Visible: true, Order: 0},
			{WidgetID: WidgetAuditeeComplianceSum, Visible: true, Order: 1},
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

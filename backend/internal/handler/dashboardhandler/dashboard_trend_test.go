package dashboardhandler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/middleware"
)

func TestBuildTrendPeriod(t *testing.T) {
	now := time.Date(2026, time.August, 27, 16, 0, 0, 0, dashboardLocation())
	tests := []struct {
		period      string
		start       string
		end         string
		bucketCount int
		granularity trendGranularity
	}{
		{"daily", "2026-08-14", "2026-08-28", 14, trendDay},
		{"weekly", "2026-06-08", "2026-08-31", 12, trendWeek},
		{"monthly", "2025-09-01", "2026-09-01", 12, trendMonth},
		{"quarter", "2024-10-01", "2026-10-01", 8, trendQuarter},
	}
	for _, test := range tests {
		t.Run(test.period, func(t *testing.T) {
			spec, err := buildTrendPeriod(test.period, now, "", "")
			if err != nil {
				t.Fatal(err)
			}
			if got := spec.Start.Format("2006-01-02"); got != test.start {
				t.Fatalf("start = %s, want %s", got, test.start)
			}
			if got := spec.End.Format("2006-01-02"); got != test.end {
				t.Fatalf("end = %s, want %s", got, test.end)
			}
			if len(spec.Buckets) != test.bucketCount || spec.Granularity != test.granularity {
				t.Fatalf("buckets/granularity = %d/%s, want %d/%s", len(spec.Buckets), spec.Granularity, test.bucketCount, test.granularity)
			}
		})
	}
}

// Custom (user-picked start_date/end_date) replaces the old
// previous_week/previous_month/previous_year single-past-period presets —
// granularity auto-scales with the picked range's length.
func TestBuildTrendPeriodCustomRange(t *testing.T) {
	now := time.Date(2026, time.August, 27, 16, 0, 0, 0, dashboardLocation())
	tests := []struct {
		name        string
		start       string
		end         string
		wantStart   string
		wantEnd     string
		granularity trendGranularity
	}{
		{"short range buckets by day", "2026-08-17", "2026-08-23", "2026-08-17", "2026-08-24", trendDay},
		{"medium range buckets by week", "2026-06-01", "2026-08-27", "2026-06-01", "2026-08-28", trendWeek},
		{"long range buckets by month", "2025-01-01", "2026-01-01", "2025-01-01", "2026-01-02", trendMonth},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := buildTrendPeriod("custom", now, test.start, test.end)
			if err != nil {
				t.Fatal(err)
			}
			if got := spec.Start.Format("2006-01-02"); got != test.wantStart {
				t.Fatalf("start = %s, want %s", got, test.wantStart)
			}
			if got := spec.End.Format("2006-01-02"); got != test.wantEnd {
				t.Fatalf("end = %s, want %s", got, test.wantEnd)
			}
			if spec.Granularity != test.granularity {
				t.Fatalf("granularity = %s, want %s", spec.Granularity, test.granularity)
			}
		})
	}
}

func TestBuildTrendPeriodCustomRangeRequiresValidDates(t *testing.T) {
	now := time.Now()
	cases := map[string][2]string{
		"missing start_date": {"", "2026-08-01"},
		"missing end_date":   {"2026-08-01", ""},
		"unparsable date":    {"08/01/2026", "2026-08-01"},
		"end before start":   {"2026-08-10", "2026-08-01"},
	}
	for name, dates := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := buildTrendPeriod("custom", now, dates[0], dates[1]); err == nil {
				t.Fatal("expected invalid custom range to be rejected")
			}
		})
	}
}

func TestBuildTrendPeriodRejectsUnknownValue(t *testing.T) {
	if _, err := buildTrendPeriod("unknown", time.Now(), "", ""); err == nil {
		t.Fatal("expected invalid period to be rejected")
	}
}

func TestBuildTrendPeriodHandlesLeapMonth(t *testing.T) {
	now := time.Date(2024, time.March, 15, 12, 0, 0, 0, dashboardLocation())
	spec, err := buildTrendPeriod("custom", now, "2024-02-01", "2024-02-29")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(spec.Buckets); got != 29 {
		t.Fatalf("bucket count = %d, want 29", got)
	}
	if got := spec.Start.Format("2006-01-02"); got != "2024-02-01" {
		t.Fatalf("start = %s, want 2024-02-01", got)
	}
	if got := spec.End.Format("2006-01-02"); got != "2024-03-01" {
		t.Fatalf("end = %s, want 2024-03-01", got)
	}
}

func TestBuildTrendPeriodSupportsLegacyAliases(t *testing.T) {
	now := time.Date(2026, time.August, 27, 16, 0, 0, 0, dashboardLocation())
	tests := map[string]string{"1m": "weekly", "3m": "monthly", "6m": "monthly", "1y": "monthly", "12m": "monthly"}
	for alias, expected := range tests {
		spec, err := buildTrendPeriod(alias, now, "", "")
		if err != nil {
			t.Fatalf("alias %s: %v", alias, err)
		}
		if spec.Period != expected {
			t.Fatalf("alias %s resolved to %s, want %s", alias, spec.Period, expected)
		}
	}
}

func TestResolveTrendInspectorScopesOnlyAuditor(t *testing.T) {
	tests := []struct {
		name      string
		roleID    string
		userID    string
		requested string
		expected  string
		wantError bool
	}{
		{name: "admin sees all by default", roleID: "ROLE-001", userID: "admin-1"},
		{name: "admin can select auditor", roleID: "ROLE-001", userID: "admin-1", requested: "auditor-2", expected: "auditor-2"},
		{name: "auditor is automatically scoped", roleID: "ROLE-002", userID: "auditor-1", expected: "auditor-1"},
		{name: "auditor cannot select another auditor", roleID: "ROLE-002", userID: "auditor-1", requested: "auditor-2", wantError: true},
		{name: "non admin cannot filter auditor", roleID: "ROLE-004", userID: "supervisor-1", requested: "auditor-2", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			var result string
			var resultErr error
			app.Get("/", func(c *fiber.Ctx) error {
				c.Locals(middleware.ContextKeyUserID, test.userID)
				c.Locals(middleware.ContextKeyRoleID, test.roleID)
				result, resultErr = resolveTrendInspector(c)
				return c.SendStatus(fiber.StatusNoContent)
			})

			target := "/"
			if test.requested != "" {
				target += "?inspector_id=" + test.requested
			}
			request := httptest.NewRequest("GET", target, nil)
			if _, err := app.Test(request); err != nil {
				t.Fatal(err)
			}
			if (resultErr != nil) != test.wantError {
				t.Fatalf("error = %v, wantError %v", resultErr, test.wantError)
			}
			if result != test.expected {
				t.Fatalf("inspector = %q, want %q", result, test.expected)
			}
		})
	}
}

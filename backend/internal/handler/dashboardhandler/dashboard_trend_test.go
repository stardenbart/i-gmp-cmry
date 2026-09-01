package dashboardhandler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/middleware"
)

// "quarter" is the one fixed rolling-window preset kept outside the
// date-range picker (per explicit product decision: quarter stays, but is
// NOT part of the calendar filter).
func TestBuildTrendPeriodQuarterPreset(t *testing.T) {
	now := time.Date(2026, time.August, 27, 16, 0, 0, 0, dashboardLocation())
	spec, err := buildTrendPeriod("quarter", now, "", "", "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := spec.Start.Format("2006-01-02"); got != "2024-10-01" {
		t.Fatalf("start = %s, want 2024-10-01", got)
	}
	if got := spec.End.Format("2006-01-02"); got != "2026-10-01" {
		t.Fatalf("end = %s, want 2026-10-01", got)
	}
	if len(spec.Buckets) != 8 || spec.Granularity != trendQuarter {
		t.Fatalf("buckets/granularity = %d/%s, want 8/%s", len(spec.Buckets), spec.Granularity, trendQuarter)
	}
}

// "range" is the everyday mode: the user always picks start_date/end_date,
// and granularity (day/week/month/year) controls how it's bucketed for the
// chart — this replaced the old daily/weekly/monthly/previous_* presets.
func TestBuildTrendPeriodRange(t *testing.T) {
	now := time.Date(2026, time.August, 27, 16, 0, 0, 0, dashboardLocation())
	tests := []struct {
		name        string
		start       string
		end         string
		granularity string
		wantStart   string
		wantEnd     string
		wantGran    trendGranularity
		wantBuckets int
	}{
		{"daily granularity", "2026-08-14", "2026-08-27", "day", "2026-08-14", "2026-08-28", trendDay, 14},
		{"weekly granularity", "2026-06-08", "2026-08-30", "week", "2026-06-08", "2026-08-31", trendWeek, 12},
		{"monthly granularity", "2025-09-01", "2026-08-31", "month", "2025-09-01", "2026-09-01", trendMonth, 12},
		{"yearly granularity", "2020-01-01", "2025-12-31", "year", "2020-01-01", "2026-01-01", trendYear, 6},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := buildTrendPeriod("range", now, test.start, test.end, test.granularity, 1)
			if err != nil {
				t.Fatal(err)
			}
			if got := spec.Start.Format("2006-01-02"); got != test.wantStart {
				t.Fatalf("start = %s, want %s", got, test.wantStart)
			}
			if got := spec.End.Format("2006-01-02"); got != test.wantEnd {
				t.Fatalf("end = %s, want %s", got, test.wantEnd)
			}
			if spec.Granularity != test.wantGran {
				t.Fatalf("granularity = %s, want %s", spec.Granularity, test.wantGran)
			}
			if len(spec.Buckets) != test.wantBuckets {
				t.Fatalf("buckets = %d, want %d", len(spec.Buckets), test.wantBuckets)
			}
		})
	}
}

func TestBuildTrendPeriodRangeRequiresValidInput(t *testing.T) {
	now := time.Now()
	cases := map[string]struct{ start, end, granularity string }{
		"missing start_date":  {"", "2026-08-01", "day"},
		"missing end_date":    {"2026-08-01", "", "day"},
		"unparsable date":     {"08/01/2026", "2026-08-10", "day"},
		"end before start":    {"2026-08-10", "2026-08-01", "day"},
		"missing granularity": {"2026-08-01", "2026-08-10", ""},
		"invalid granularity": {"2026-08-01", "2026-08-10", "quarter"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := buildTrendPeriod("range", now, c.start, c.end, c.granularity, 1); err == nil {
				t.Fatal("expected invalid range input to be rejected")
			}
		})
	}
}

// Picking a fine granularity across a huge range would produce thousands of
// unreadable chart points — must be rejected with a clear message instead.
func TestBuildTrendPeriodRangeRejectsTooManyBuckets(t *testing.T) {
	now := time.Now()
	_, err := buildTrendPeriod("range", now, "2015-01-01", "2025-12-31", "day", 1)
	if err == nil {
		t.Fatal("expected a multi-year daily range to be rejected")
	}
}

func TestBuildTrendPeriodRejectsUnknownValue(t *testing.T) {
	if _, err := buildTrendPeriod("unknown", time.Now(), "", "", "", 1); err == nil {
		t.Fatal("expected invalid period to be rejected")
	}
}

func TestBuildTrendPeriodHandlesLeapMonth(t *testing.T) {
	now := time.Date(2024, time.March, 15, 12, 0, 0, 0, dashboardLocation())
	spec, err := buildTrendPeriod("range", now, "2024-02-01", "2024-02-29", "day", 1)
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

// A non-default cutoff day aligns "month" buckets to the same inspection
// period used for Kawasan/Area completion tracking, instead of the raw
// picked start_date or the 1st of the calendar month.
func TestBuildTrendPeriodMonthGranularityRespectsCutoffDay(t *testing.T) {
	now := time.Date(2026, time.August, 27, 16, 0, 0, 0, dashboardLocation())
	// User picks 20 Jan-20 Mar with cutoff day 13: the first bucket must
	// extend back to 13 Jan (the period start containing 20 Jan), not stay
	// at the picked 20 Jan — otherwise "Jan" would be a meaningless partial
	// month that doesn't match any real inspection period.
	spec, err := buildTrendPeriod("range", now, "2026-01-20", "2026-03-20", "month", 13)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Buckets) < 1 {
		t.Fatal("expected at least one bucket")
	}
	first := spec.Buckets[0]
	if got := first.Start.Format("2006-01-02"); got != "2026-01-13" {
		t.Fatalf("first bucket start = %s, want 2026-01-13", got)
	}
	if got := first.Key; got != "2026-01" {
		t.Fatalf("first bucket key = %s, want 2026-01", got)
	}
	if got := first.Label; got != "Jan 2026" {
		t.Fatalf("first bucket label = %s, want Jan 2026", got)
	}
	second := spec.Buckets[1]
	if got := second.Start.Format("2006-01-02"); got != "2026-02-13" {
		t.Fatalf("second bucket start = %s, want 2026-02-13", got)
	}
	if got := second.Key; got != "2026-02" {
		t.Fatalf("second bucket key = %s, want 2026-02", got)
	}
}

// cutoffDay=1 (the default) must produce identical buckets to the old,
// pre-cutoff behavior — no regression for plants that never change the
// setting.
func TestBuildTrendPeriodMonthGranularityCutoffDayOneIsCalendarMonth(t *testing.T) {
	now := time.Date(2026, time.August, 27, 16, 0, 0, 0, dashboardLocation())
	spec, err := buildTrendPeriod("range", now, "2026-01-01", "2026-03-01", "month", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := spec.Buckets[0].Start.Format("2006-01-02"); got != "2026-01-01" {
		t.Fatalf("first bucket start = %s, want 2026-01-01", got)
	}
}

func TestTrendBucketSQLMonthShiftsByCutoffDayMinusOne(t *testing.T) {
	if got := trendBucketSQL(`"col"`, trendMonth, 1); got != `TO_CHAR("col", 'YYYY-MM')` {
		t.Fatalf("cutoffDay=1: got %q, want no shift", got)
	}
	got := trendBucketSQL(`"col"`, trendMonth, 13)
	want := `TO_CHAR("col" - INTERVAL '12 days', 'YYYY-MM')`
	if got != want {
		t.Fatalf("cutoffDay=13: got %q, want %q", got, want)
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

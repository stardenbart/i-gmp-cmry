package dashboardhandler

import (
	"testing"
	"time"
)

func TestParseStatsDateRangeEmptyMeansAllTime(t *testing.T) {
	r, err := parseStatsDateRange("", "", dashboardLocation())
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Fatalf("range = %+v, want nil (no filter)", r)
	}
}

func TestParseStatsDateRangeIsInclusiveWIBDays(t *testing.T) {
	loc := dashboardLocation()
	r, err := parseStatsDateRange("2026-10-01", "2026-10-08", loc)
	if err != nil {
		t.Fatal(err)
	}
	wantFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	wantTo := time.Date(2026, 10, 9, 0, 0, 0, 0, loc) // exclusive: all of 8 Oct counts
	if r.From == nil || !r.From.Equal(wantFrom) {
		t.Fatalf("from = %v, want %v", r.From, wantFrom)
	}
	if r.To == nil || !r.To.Equal(wantTo) {
		t.Fatalf("to = %v, want %v", r.To, wantTo)
	}
}

func TestParseStatsDateRangeAllowsOpenEnds(t *testing.T) {
	loc := dashboardLocation()
	r, err := parseStatsDateRange("2026-10-01", "", loc)
	if err != nil {
		t.Fatal(err)
	}
	if r.From == nil || r.To != nil {
		t.Fatalf("start only: from=%v to=%v, want from set, to nil", r.From, r.To)
	}

	r, err = parseStatsDateRange("", "2026-10-08", loc)
	if err != nil {
		t.Fatal(err)
	}
	if r.From != nil || r.To == nil {
		t.Fatalf("end only: from=%v to=%v, want from nil, to set", r.From, r.To)
	}
}

func TestParseStatsDateRangeRejectsBadInput(t *testing.T) {
	loc := dashboardLocation()
	for _, tc := range []struct{ name, start, end string }{
		{"bad start format", "01-10-2026", ""},
		{"bad end format", "", "2026/10/08"},
		{"end before start", "2026-10-08", "2026-10-01"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseStatsDateRange(tc.start, tc.end, loc); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestStatsDateRangeSQLCondition(t *testing.T) {
	loc := dashboardLocation()
	r, _ := parseStatsDateRange("2026-10-01", "2026-10-08", loc)
	sql, args := r.condition(`ih."InspectionHeaderCreatedAt"`)
	if sql != `ih."InspectionHeaderCreatedAt" >= ? AND ih."InspectionHeaderCreatedAt" < ?` {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 {
		t.Fatalf("args = %v, want 2", args)
	}

	open, _ := parseStatsDateRange("2026-10-01", "", loc)
	sql, args = open.condition(`ih."InspectionHeaderCreatedAt"`)
	if sql != `ih."InspectionHeaderCreatedAt" >= ?` || len(args) != 1 {
		t.Fatalf("open-ended sql = %q args = %v", sql, args)
	}
}

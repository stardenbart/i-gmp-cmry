package inspectionusecase

import (
	"testing"
	"time"
)

func TestResolveInspectionPeriod_CutoffDayOneMatchesCalendarMonth(t *testing.T) {
	now := time.Date(2026, time.January, 20, 15, 30, 0, 0, time.UTC)
	start, end := ResolveInspectionPeriod(now, 1)

	wantStart := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("got [%v, %v), want [%v, %v)", start, end, wantStart, wantEnd)
	}
}

func TestResolveInspectionPeriod_CutoffDayThirteenBeforeCutoff(t *testing.T) {
	// 5 Jan, cutoff=13 → still inside the period that began 13 Dec (prev year).
	now := time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)
	start, end := ResolveInspectionPeriod(now, 13)

	wantStart := time.Date(2025, time.December, 13, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.January, 13, 0, 0, 0, 0, time.UTC)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("got [%v, %v), want [%v, %v)", start, end, wantStart, wantEnd)
	}
}

func TestResolveInspectionPeriod_CutoffDayThirteenOnAndAfterCutoff(t *testing.T) {
	cases := []time.Time{
		time.Date(2026, time.January, 13, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.February, 12, 23, 59, 59, 0, time.UTC),
	}
	wantStart := time.Date(2026, time.January, 13, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.February, 13, 0, 0, 0, 0, time.UTC)

	for _, now := range cases {
		start, end := ResolveInspectionPeriod(now, 13)
		if !start.Equal(wantStart) || !end.Equal(wantEnd) {
			t.Fatalf("for t=%v: got [%v, %v), want [%v, %v)", now, start, end, wantStart, wantEnd)
		}
	}
}

func TestResolveInspectionPeriod_InvalidCutoffDayFallsBackToDefault(t *testing.T) {
	now := time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
	for _, invalid := range []int{0, -5, 29, 31, 100} {
		start, end := ResolveInspectionPeriod(now, invalid)
		wantStart := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		wantEnd := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
		if !start.Equal(wantStart) || !end.Equal(wantEnd) {
			t.Fatalf("for cutoffDay=%d: got [%v, %v), want default [%v, %v)", invalid, start, end, wantStart, wantEnd)
		}
	}
}

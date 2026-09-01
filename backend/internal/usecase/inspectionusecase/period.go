package inspectionusecase

import "time"

// DefaultInspectionPeriodCutoffDay preserves exact current behavior (a
// plain calendar month) for any plant that hasn't changed the setting.
const DefaultInspectionPeriodCutoffDay = 1

// ResolveInspectionPeriod returns the [start, end) boundaries of the
// "inspection period" that `t` falls into, given a configurable cutoff day
// (the day of month a new period begins). cutoffDay=1 is an exact calendar
// month. Example: cutoffDay=13 means 13 Jan 00:00-13 Feb 00:00 is "the
// January period" — a DetailKawasan inspected anywhere in that window
// counts toward January, not February.
func ResolveInspectionPeriod(t time.Time, cutoffDay int) (start, end time.Time) {
	if cutoffDay < 1 || cutoffDay > 28 {
		cutoffDay = DefaultInspectionPeriodCutoffDay
	}
	year, month, day := t.Date()
	if day < cutoffDay {
		// Belongs to the period that began last month. time.Date normalizes
		// month=0 to December of the previous year automatically, so no
		// manual year-rollover logic is needed here.
		month--
	}
	start = time.Date(year, month, cutoffDay, 0, 0, 0, 0, t.Location())
	return start, start.AddDate(0, 1, 0)
}

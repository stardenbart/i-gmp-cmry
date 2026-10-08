package dashboardhandler

import (
	"errors"
	"fmt"
	"time"
)

// statsDateRange is the overview filter for /dashboard/stats: whole WIB
// calendar days, From inclusive and To exclusive (the day after end_date).
// Either side may be nil for an open-ended range.
type statsDateRange struct {
	From *time.Time
	To   *time.Time
}

// parseStatsDateRange reads start_date/end_date (YYYY-MM-DD). Both empty
// means no filter (all time) and returns nil.
func parseStatsDateRange(start, end string, loc *time.Location) (*statsDateRange, error) {
	if start == "" && end == "" {
		return nil, nil
	}
	r := &statsDateRange{}
	if start != "" {
		s, err := time.ParseInLocation("2006-01-02", start, loc)
		if err != nil {
			return nil, fmt.Errorf("start_date harus berformat YYYY-MM-DD")
		}
		r.From = &s
	}
	if end != "" {
		e, err := time.ParseInLocation("2006-01-02", end, loc)
		if err != nil {
			return nil, fmt.Errorf("end_date harus berformat YYYY-MM-DD")
		}
		next := e.AddDate(0, 0, 1)
		r.To = &next
	}
	if r.From != nil && r.To != nil && !r.To.After(*r.From) {
		return nil, errors.New("end_date tidak boleh sebelum start_date")
	}
	return r, nil
}

// condition returns the WHERE fragment and args that keep column inside the
// range.
func (r *statsDateRange) condition(column string) (string, []interface{}) {
	switch {
	case r.From != nil && r.To != nil:
		return column + " >= ? AND " + column + " < ?", []interface{}{*r.From, *r.To}
	case r.From != nil:
		return column + " >= ?", []interface{}{*r.From}
	default:
		return column + " < ?", []interface{}{*r.To}
	}
}

// wowrConditions are WHERE fragments over "Issue" i for the overview's WO/WR
// buckets. They mirror the follow-up page: PendingValidation counts as
// waiting for the auditor only once a WO/WR proof photo exists.
type wowrConditions struct {
	Pending  string
	Awaiting string
}

const wowrHasProofSQL = `EXISTS (SELECT 1 FROM "Issue_Photo" proof WHERE proof."IssueID" = i."IssueID" AND proof."PhotoType" = 'WOWR' AND COALESCE(proof."ImageUrl", '') != '')`

func wowrStatusConditions() wowrConditions {
	return wowrConditions{
		Pending:  `i."WOWRStatus" = 'PendingValidation' AND ` + wowrHasProofSQL,
		Awaiting: `(i."WOWRStatus" = 'None' OR i."WOWRStatus" IS NULL OR (i."WOWRStatus" = 'PendingValidation' AND NOT ` + wowrHasProofSQL + `))`,
	}
}

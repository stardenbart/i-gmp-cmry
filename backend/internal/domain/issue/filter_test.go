package issue

import (
	"testing"
	"time"
)

func TestParseIssueDateUsesJakartaDayBoundary(t *testing.T) {
	from, err := parseIssueDate("2026-09-01", false)
	if err != nil {
		t.Fatalf("parse date_from: %v", err)
	}
	to, err := parseIssueDate("2026-09-01", true)
	if err != nil {
		t.Fatalf("parse date_to: %v", err)
	}

	if got := from.Format(time.RFC3339); got != "2026-09-01T00:00:00+07:00" {
		t.Fatalf("unexpected date_from boundary: %s", got)
	}
	if got := to.Format(time.RFC3339); got != "2026-09-02T00:00:00+07:00" {
		t.Fatalf("date_to must be next-day exclusive boundary, got %s", got)
	}
}

func TestParseIssueDateRejectsInvalidValue(t *testing.T) {
	if _, err := parseIssueDate("01/09/2026", false); err == nil {
		t.Fatal("expected invalid date format to return an error")
	}
}

func TestOpenOverdueStatusConditionIncludesOpenAndInProgress(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	query, args := issueStatusCondition(string(IssueStatusOpenOverdue), now)

	if query == "" || len(args) != 2 {
		t.Fatalf("unexpected condition: query=%q args=%v", query, args)
	}
	statuses, ok := args[0].([]IssueStatus)
	if !ok || len(statuses) != 2 || statuses[0] != IssueStatusOpen || statuses[1] != IssueStatusInProgress {
		t.Fatalf("unexpected overdue source statuses: %#v", args[0])
	}
}

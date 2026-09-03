package issueusecase

import (
	"testing"
	"time"

	"github.com/monitoring-system/backend/internal/domain/issue"
)

type issueFilterRepositoryStub struct {
	exportItems []issue.Issue
}

func (s *issueFilterRepositoryStub) FindFiltered(*issue.IssueFilter) ([]issue.Issue, int64, error) {
	return nil, 0, nil
}

func (s *issueFilterRepositoryStub) FindForExport(*issue.IssueFilter, int) ([]issue.Issue, error) {
	return s.exportItems, nil
}

func (s *issueFilterRepositoryStub) FindFacets(*issue.IssueFilter) (issue.IssueFacets, error) {
	return issue.IssueFacets{}, nil
}

func TestGetForExportRejectsRowsAboveLimit(t *testing.T) {
	repo := &issueFilterRepositoryStub{exportItems: make([]issue.Issue, 3)}
	uc := NewIssueFilterUseCase(repo, nil)

	if _, err := uc.GetForExport(&issue.IssueFilter{}, 2); err == nil {
		t.Fatal("expected export limit error")
	}
}

func TestGetForExportSetsComputedStatus(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	due := now.Add(-time.Hour)
	repo := &issueFilterRepositoryStub{exportItems: []issue.Issue{{
		IssueID:     "ISS-001",
		IssueStatus: issue.IssueStatusOpen,
		DueDate:     &due,
	}}}
	uc := NewIssueFilterUseCase(repo, nil)

	items, err := uc.GetForExport(&issue.IssueFilter{Now: now}, 2)
	if err != nil {
		t.Fatalf("GetForExport: %v", err)
	}
	if got := items[0].ComputedIssueStatus; got != issue.IssueStatusOpenOverdue {
		t.Fatalf("unexpected computed status: %s", got)
	}
}

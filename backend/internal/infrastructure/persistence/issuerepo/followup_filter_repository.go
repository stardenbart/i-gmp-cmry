package issuerepo

import (
	"time"

	"github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
)

// ── Followup Filter Repository ────────────────────────────────────────────

type followupFilterRepository struct{ db *gorm.DB }

// NewFollowupFilterRepository creates a new FollowupFilterRepository.
func NewFollowupFilterRepository(db *gorm.DB) issue.FollowupFilterRepository {
	return &followupFilterRepository{db: db}
}

func (r *followupFilterRepository) FindFiltered(f *issue.FollowupFilter) ([]issue.Issue, int64, error) {
	var items []issue.Issue
	var total int64

	base := r.db.Model(&issue.Issue{})
	base = f.ApplyTo(base)
	base.Count(&total)

	err := f.ApplySort(base).
		Preload("Photos").
		Offset((f.Page - 1) * f.Limit).
		Limit(f.Limit).
		Find(&items).Error

	return items, total, err
}

func (r *followupFilterRepository) FindFacets(f *issue.FollowupFilter) (issue.FollowupFacets, error) {
	facets := issue.FollowupFacets{
		Status:     make(map[string]int64),
		WOWRStatus: make(map[string]int64),
		DateRange: issue.DateRangeFacet{
			Field: "created_at",
		},
		DueDateRange: issue.DateRangeFacet{
			Field: "due_date",
		},
	}

	type kv struct {
		Key   string
		Count int64
	}

	// Helper: build base query excluding a specific field
	baseWithout := func(excludeStatus, excludeWOWR, excludeDate, excludeDue bool) *gorm.DB {
		tmp := &issue.FollowupFilter{
			Q:             f.Q,
			ScopeUserID:   f.ScopeUserID,
			NeedsWOWR:     f.NeedsWOWR,
			Label:         f.Label,
			IncludeClosed: f.IncludeClosed,
		}
		if !excludeStatus {
			tmp.Status = f.Status
			tmp.StatusIn = f.StatusIn
		}
		if !excludeWOWR {
			tmp.WOWRStatus = f.WOWRStatus
			tmp.WOWRStatusIn = f.WOWRStatusIn
		}
		if !excludeDate {
			tmp.DateFrom = f.DateFrom
			tmp.DateTo = f.DateTo
		}
		if !excludeDue {
			tmp.DueFrom = f.DueFrom
			tmp.DueTo = f.DueTo
		}
		return tmp.ApplyTo(r.db.Model(&issue.Issue{}))
	}

	// Status facet
	var statusRows []kv
	if err := baseWithout(true, false, false, false).
		Select(`"Issue"."IssueStatus" AS key, COUNT(*) AS count`).
		Group(`"Issue"."IssueStatus"`).
		Scan(&statusRows).Error; err == nil {
		for _, row := range statusRows {
			facets.Status[row.Key] = row.Count
		}
	}

	// WOWRStatus facet
	var wowrRows []kv
	if err := baseWithout(false, true, false, false).
		Select(`"Issue"."WOWRStatus" AS key, COUNT(*) AS count`).
		Group(`"Issue"."WOWRStatus"`).
		Scan(&wowrRows).Error; err == nil {
		for _, row := range wowrRows {
			facets.WOWRStatus[row.Key] = row.Count
		}
	}

	// Date range (created_at)
	type dateResult struct {
		Min *time.Time
		Max *time.Time
	}
	var dr dateResult
	if err := baseWithout(false, false, true, false).
		Select(`MIN("Issue"."IssueCreatedAt") AS min, MAX("Issue"."IssueCreatedAt") AS max`).
		Scan(&dr).Error; err == nil {
		facets.DateRange.Min = dr.Min
		facets.DateRange.Max = dr.Max
	}

	// DueDate range
	var ddr dateResult
	if err := baseWithout(false, false, false, true).
		Select(`MIN("Issue"."DueDate") AS min, MAX("Issue"."DueDate") AS max`).
		Scan(&ddr).Error; err == nil {
		facets.DueDateRange.Min = ddr.Min
		facets.DueDateRange.Max = ddr.Max
	}

	return facets, nil
}

package issuerepo

import (
	"time"

	"github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
)

// ── Issue Filter Repository ───────────────────────────────────────────────

type issueFilterRepository struct{ db *gorm.DB }

// NewIssueFilterRepository creates a new IssueFilterRepository.
func NewIssueFilterRepository(db *gorm.DB) issue.IssueFilterRepository {
	return &issueFilterRepository{db: db}
}

func (r *issueFilterRepository) FindFiltered(f *issue.IssueFilter) ([]issue.Issue, int64, error) {
	var items []issue.Issue
	var total int64

	base := r.db.Model(&issue.Issue{})
	base = f.ApplyTo(base)
	base.Count(&total)

	err := f.ApplySort(base).
		Select(`"Issue".*, 
			am."AreaName" AS "AreaName", 
			km."KawasanName" AS "KawasanName", 
			dkm."DetailKawasanName" AS "DetailKawasanName", 
			u."FullName" AS "PICName"`).
		Joins(`LEFT JOIN "Inspection_Result" ir ON ir."ResultID" = "Issue"."ResultID"`).
		Joins(`LEFT JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Joins(`LEFT JOIN "Area_Master" am ON am."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = "Issue"."IssuePICUserID"`).
		Preload("Photos").
		Offset((f.Page - 1) * f.Limit).
		Limit(f.Limit).
		Find(&items).Error

	return items, total, err
}

func (r *issueFilterRepository) FindFacets(f *issue.IssueFilter) (issue.IssueFacets, error) {
	facets := issue.IssueFacets{
		Status:     make(map[string]int64),
		WOWRStatus: make(map[string]int64),
		PICUserID:  make(map[string]int64),
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

	// Helper: build base query excluding a particular field's filter
	baseWithout := func(excludeStatus, excludePIC, excludeWOWR, excludeDate, excludeDue bool) *gorm.DB {
		tmp := &issue.IssueFilter{
			Q:           f.Q,
			ScopeUserID: f.ScopeUserID,
			NeedsWOWR:   f.NeedsWOWR,
			Label:       f.Label,
		}
		if !excludeStatus {
			tmp.Status = f.Status
			tmp.StatusIn = f.StatusIn
		}
		if !excludePIC {
			tmp.PICUserID = f.PICUserID
			tmp.PICUserIDIn = f.PICUserIDIn
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
	if err := baseWithout(true, false, false, false, false).
		Select(`"IssueStatus" AS key, COUNT(*) AS count`).
		Group(`"IssueStatus"`).
		Scan(&statusRows).Error; err == nil {
		for _, row := range statusRows {
			facets.Status[row.Key] = row.Count
		}
	}

	// PICUserID facet
	var picRows []kv
	if err := baseWithout(false, true, false, false, false).
		Select(`"IssuePICUserID" AS key, COUNT(*) AS count`).
		Group(`"IssuePICUserID"`).
		Scan(&picRows).Error; err == nil {
		for _, row := range picRows {
			facets.PICUserID[row.Key] = row.Count
		}
	}

	// WOWRStatus facet
	var wowrRows []kv
	if err := baseWithout(false, false, true, false, false).
		Select(`"WOWRStatus" AS key, COUNT(*) AS count`).
		Group(`"WOWRStatus"`).
		Scan(&wowrRows).Error; err == nil {
		for _, row := range wowrRows {
			facets.WOWRStatus[row.Key] = row.Count
		}
	}

	// Date range (created_at, without date filter)
	type dateResult struct {
		Min *time.Time
		Max *time.Time
	}
	var dr dateResult
	if err := baseWithout(false, false, false, true, false).
		Select(`MIN("IssueCreatedAt") AS min, MAX("IssueCreatedAt") AS max`).
		Scan(&dr).Error; err == nil {
		facets.DateRange.Min = dr.Min
		facets.DateRange.Max = dr.Max
	}

	// DueDate range (without due filter)
	var ddr dateResult
	if err := baseWithout(false, false, false, false, true).
		Select(`MIN("DueDate") AS min, MAX("DueDate") AS max`).
		Scan(&ddr).Error; err == nil {
		facets.DueDateRange.Min = ddr.Min
		facets.DueDateRange.Max = ddr.Max
	}

	return facets, nil
}

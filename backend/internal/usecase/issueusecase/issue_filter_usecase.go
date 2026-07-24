package issueusecase

import (
	"math"
	"sync"

	"github.com/monitoring-system/backend/internal/domain/issue"
)

type issueFilterUseCase struct {
	repo issue.IssueFilterRepository
}

// NewIssueFilterUseCase creates a new IssueFilterUseCase.
func NewIssueFilterUseCase(repo issue.IssueFilterRepository) issue.IssueFilterUseCase {
	return &issueFilterUseCase{repo: repo}
}

// GetFiltered fetches filtered+paginated issues and facets in parallel.
func (uc *issueFilterUseCase) GetFiltered(f *issue.IssueFilter) (*issue.IssueFilterResult, error) {
	var (
		items     []issue.Issue
		total     int64
		facets    issue.IssueFacets
		errItems  error
		errFacets error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		items, total, errItems = uc.repo.FindFiltered(f)
	}()

	go func() {
		defer wg.Done()
		facets, errFacets = uc.repo.FindFacets(f)
	}()

	wg.Wait()

	if errItems != nil {
		return nil, errItems
	}
	if errFacets != nil {
		return nil, errFacets
	}

	totalPages := int64(math.Ceil(float64(total) / float64(f.Limit)))

	return &issue.IssueFilterResult{
		Items:          items,
		Total:          total,
		Page:           f.Page,
		Limit:          f.Limit,
		TotalPages:     totalPages,
		FiltersApplied: f.FiltersApplied(),
		Facets:         facets,
	}, nil
}

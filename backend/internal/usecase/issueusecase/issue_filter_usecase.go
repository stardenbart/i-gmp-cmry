package issueusecase

import (
	"math"
	"sync"

	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/pkg/crypto"
)

type issueFilterUseCase struct {
	repo      issue.IssueFilterRepository
	cryptoSvc *crypto.Service
}

// NewIssueFilterUseCase creates a new IssueFilterUseCase.
func NewIssueFilterUseCase(repo issue.IssueFilterRepository, cryptoSvc *crypto.Service) issue.IssueFilterUseCase {
	return &issueFilterUseCase{repo: repo, cryptoSvc: cryptoSvc}
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

	if uc.cryptoSvc != nil {
		for i := range items {
			items[i].Keterangan = uc.cryptoSvc.DecryptWithFallback(items[i].Keterangan)
			for j := range items[i].Photos {
				items[i].Photos[j].ImageUrl = uc.cryptoSvc.DecryptWithFallback(items[i].Photos[j].ImageUrl)
				items[i].Photos[j].FileName = uc.cryptoSvc.DecryptWithFallback(items[i].Photos[j].FileName)
			}
		}
	}

	// Deduplicate items by IssueID to prevent SQL join duplication
	seen := make(map[string]bool)
	unique := make([]issue.Issue, 0, len(items))
	for _, item := range items {
		if item.IssueID != "" && !seen[item.IssueID] {
			seen[item.IssueID] = true
			unique = append(unique, item)
		}
	}
	items = unique

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

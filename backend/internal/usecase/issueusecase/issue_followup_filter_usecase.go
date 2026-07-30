package issueusecase

import (
	"math"
	"sync"

	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/pkg/crypto"
)

type followupFilterUseCase struct {
	repo      issue.FollowupFilterRepository
	cryptoSvc *crypto.Service
}

// NewFollowupFilterUseCase creates a new FollowupFilterUseCase.
func NewFollowupFilterUseCase(repo issue.FollowupFilterRepository, cryptoSvc *crypto.Service) issue.FollowupFilterUseCase {
	return &followupFilterUseCase{repo: repo, cryptoSvc: cryptoSvc}
}

// GetFiltered fetches followup issues and facets in parallel.
func (uc *followupFilterUseCase) GetFiltered(f *issue.FollowupFilter) (*issue.FollowupFilterResult, error) {
	var (
		items     []issue.Issue
		total     int64
		facets    issue.FollowupFacets
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

	totalPages := int64(math.Ceil(float64(total) / float64(f.Limit)))

	return &issue.FollowupFilterResult{
		Items:          items,
		Total:          total,
		Page:           f.Page,
		Limit:          f.Limit,
		TotalPages:     totalPages,
		FiltersApplied: f.FiltersApplied(),
		Facets:         facets,
	}, nil
}

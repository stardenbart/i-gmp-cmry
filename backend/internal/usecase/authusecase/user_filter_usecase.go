package authusecase

import (
	"math"
	"sync"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
)

type userFilterUseCase struct {
	repo authdomain.UserFilterRepository
}

// NewUserFilterUseCase creates a new UserFilterUseCase.
func NewUserFilterUseCase(repo authdomain.UserFilterRepository) authdomain.UserFilterUseCase {
	return &userFilterUseCase{repo: repo}
}

// GetFiltered fetches filtered users and facets in parallel.
func (uc *userFilterUseCase) GetFiltered(f *authdomain.UserFilter) (*authdomain.UserFilterResult, error) {
	var (
		items     []authdomain.User
		total     int64
		facets    authdomain.UserFacets
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

	return &authdomain.UserFilterResult{
		Items:          items,
		Total:          total,
		Page:           f.Page,
		Limit:          f.Limit,
		TotalPages:     totalPages,
		FiltersApplied: f.FiltersApplied(),
		Facets:         facets,
	}, nil
}

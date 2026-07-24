package inspectionusecase

import (
	"math"
	"sync"

	"github.com/monitoring-system/backend/internal/domain/inspection"
)

type inspectionFilterUseCase struct {
	repo inspection.InspectionFilterRepository
}

// NewInspectionFilterUseCase creates a new InspectionFilterUseCase.
func NewInspectionFilterUseCase(repo inspection.InspectionFilterRepository) inspection.InspectionFilterUseCase {
	return &inspectionFilterUseCase{repo: repo}
}

// GetFiltered fetches filtered+paginated items and facets in parallel.
func (uc *inspectionFilterUseCase) GetFiltered(f *inspection.InspectionFilter) (*inspection.InspectionFilterResult, error) {
	var (
		items  []inspection.InspectionHeader
		total  int64
		facets inspection.InspectionFacets
		errItems error
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

	return &inspection.InspectionFilterResult{
		Items:          items,
		Total:          total,
		Page:           f.Page,
		Limit:          f.Limit,
		TotalPages:     totalPages,
		FiltersApplied: f.FiltersApplied(),
		Facets:         facets,
	}, nil
}

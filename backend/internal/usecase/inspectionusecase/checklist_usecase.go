package inspectionusecase

import (
	"context"
	"errors"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"golang.org/x/sync/singleflight"
)

// checklistSFGroup mendeduplikasi query DB GetChecklist untuk inspectionID yang sama
// yang datang secara bersamaan. TIDAK menggunakan persistent cache —
// setiap inflight window menghasilkan data fresh dari DB, tanpa risiko data basi.
var checklistSFGroup singleflight.Group

func (uc *inspectionHeaderUseCase) GetChecklist(id string) (*inspection.FullChecklist, error) {
	v, err, _ := checklistSFGroup.Do(id, func() (interface{}, error) {
		// Budget 3 detik tunggal untuk seluruh operasi (FindByID + GetFullChecklist)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		header, err := uc.repo.FindByIDWithCtx(ctx, id)
		if err != nil {
			return nil, err
		}
		if header == nil {
			return nil, errors.New("inspection not found")
		}

		return uc.repo.GetFullChecklistWithCtx(ctx, header.AreaID, id)
	})
	if err != nil {
		return nil, err
	}
	return v.(*inspection.FullChecklist), nil
}

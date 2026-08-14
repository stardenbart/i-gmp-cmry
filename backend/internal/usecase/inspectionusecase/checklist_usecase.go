package inspectionusecase

import (
	"errors"

	"github.com/monitoring-system/backend/internal/domain/inspection"
)

func (uc *inspectionHeaderUseCase) GetChecklist(id string) (*inspection.FullChecklist, error) {
	// Fetch the inspection header to get the AreaID
	header, err := uc.repo.FindByID(id)
	if err != nil {
		// Teruskan error asli ke handler, BUKAN selalu 404.
		// Sebelumnya semua error (DB timeout, pool exhaustion) disamarkan jadi "inspection not found"
		// yang menyebabkan load test melaporkan 404 padahal sebenarnya 503/500.
		return nil, err
	}
	if header == nil {
		return nil, errors.New("inspection not found")
	}

	return uc.repo.GetFullChecklist(header.AreaID, id)
}

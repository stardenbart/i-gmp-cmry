package inspectionusecase

import (
	"errors"

	"github.com/monitoring-system/backend/internal/domain/inspection"
)

func (uc *inspectionHeaderUseCase) GetChecklist(id string) (*inspection.FullChecklist, error) {
	// First fetch the inspection header to get the AreaID
	header, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("inspection not found")
	}

	return uc.repo.GetFullChecklist(header.AreaID, id)
}

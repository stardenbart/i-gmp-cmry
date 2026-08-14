package inspectionusecase

import (
	"context"
	"errors"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
)

func (uc *inspectionHeaderUseCase) GetChecklist(id string) (*inspection.FullChecklist, error) {
	// Budget 3 detik tunggal untuk seluruh operasi GetChecklist (FindByID + GetFullChecklist)
	// Mencegah akumulasi budget 3s + 3s = 6s saat pool connection sedang sibuk.
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
}

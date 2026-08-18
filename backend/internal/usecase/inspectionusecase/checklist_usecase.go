package inspectionusecase

import (
	"context"
	"errors"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
)

func (uc *inspectionHeaderUseCase) GetChecklist(id string) (*inspection.FullChecklist, error) {
	// Budget 5 detik (naik dari 3s) — mencakup 2 sequential DB ops:
	//   1. FindByIDWithCtx  → header JOIN 4 tabel + score query
	//   2. GetFullChecklistWithCtx → aspek Preload Details.Urains + InspectionResult query
	//
	// Repo layer sudah punya singleflight + TTL cache sendiri (globalFullChecklistCache 2 menit,
	// globalAspekChecklistCache 1 jam). Menambah singleflight kedua di usecase menciptakan
	// serial blocking dengan shared deadline — justru memperbesar risiko timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

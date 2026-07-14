package picusecase

import (
	"github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/idgen"
)

type picMappingUseCase struct{ repo pic.PICMappingRepository }

func NewPICMappingUseCase(repo pic.PICMappingRepository) pic.PICMappingUseCase {
	return &picMappingUseCase{repo: repo}
}

func (uc *picMappingUseCase) GetAll(page, limit int, areaID, kawasanID string) ([]pic.PICMapping, int64, error) {
	return uc.repo.FindAll(page, limit, areaID, kawasanID)
}
func (uc *picMappingUseCase) GetByID(id string) (*pic.PICMapping, error) { return uc.repo.FindByID(id) }
func (uc *picMappingUseCase) GetByUserID(userID string) ([]pic.PICMapping, error) { return uc.repo.FindByUserID(userID) }

func (uc *picMappingUseCase) Create(req *pic.CreatePICMappingRequest) (*pic.PICMapping, error) {
	item := &pic.PICMapping{
		PICMapID:    idgen.Generate(idgen.PrefixPICMap),
		AreaID:      req.AreaID,
		KawasanID:   req.KawasanID,
		UserID:      req.UserID,
		KategoriPIC: req.KategoriPIC,
	}
	return item, uc.repo.Create(item)
}

func (uc *picMappingUseCase) Update(id string, req *pic.UpdatePICMappingRequest) (*pic.PICMapping, error) {
	item, err := uc.repo.FindByID(id)
	if err != nil { return nil, err }
	if req.UserID != "" { item.UserID = req.UserID }
	if req.KategoriPIC != "" { item.KategoriPIC = req.KategoriPIC }
	return item, uc.repo.Update(item)
}

func (uc *picMappingUseCase) Delete(id string) error { return uc.repo.Delete(id) }

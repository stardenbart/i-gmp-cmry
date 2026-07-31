package masterusecase

import (
	"errors"

	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/idgen"
)

type plantUseCase struct {
	repo master.PlantRepository
}

func NewPlantUseCase(repo master.PlantRepository) master.PlantUseCase {
	return &plantUseCase{repo: repo}
}

func (u *plantUseCase) GetAll(page, limit int, search string) ([]master.Plant, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.repo.FindAll(page, limit, search)
}

func (u *plantUseCase) GetByID(id string) (*master.Plant, error) {
	return u.repo.FindByID(id)
}

func (u *plantUseCase) Create(req *master.CreatePlantRequest) (*master.Plant, error) {
	existing, _ := u.repo.FindByCode(req.PlantCode)
	if existing != nil {
		return nil, errors.New("kode plant sudah digunakan")
	}

	plant := &master.Plant{
		PlantID:   idgen.GenerateRandom("PLT"),
		PlantCode: req.PlantCode,
		PlantName: req.PlantName,
		Address:   req.Address,
	}

	if err := u.repo.Create(plant); err != nil {
		return nil, err
	}
	return plant, nil
}

func (u *plantUseCase) Update(id string, req *master.UpdatePlantRequest) (*master.Plant, error) {
	plant, err := u.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("plant tidak ditemukan")
	}

	if req.PlantCode != "" && req.PlantCode != plant.PlantCode {
		existing, _ := u.repo.FindByCode(req.PlantCode)
		if existing != nil {
			return nil, errors.New("kode plant sudah digunakan")
		}
		plant.PlantCode = req.PlantCode
	}

	if req.PlantName != "" {
		plant.PlantName = req.PlantName
	}
	plant.Address = req.Address

	if err := u.repo.Update(plant); err != nil {
		return nil, err
	}
	return plant, nil
}

func (u *plantUseCase) Delete(id string) error {
	return u.repo.Delete(id)
}

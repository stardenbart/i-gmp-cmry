package masterusecase

import (
	"context"
	"errors"
	"strings"

	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/idgen"
)

type heiUseCase struct {
	repo master.HEIRepository
}

func NewHEIUseCase(repo master.HEIRepository) master.HEIUseCase {
	return &heiUseCase{repo: repo}
}

func (uc *heiUseCase) GetAll(page, limit int, category, search string) ([]master.HEIMaster, int64, error) {
	return uc.repo.FindAll(page, limit, category, search)
}

func (uc *heiUseCase) GetByID(id string) (*master.HEIMaster, error) {
	return uc.repo.FindByID(id)
}

func (uc *heiUseCase) GetCategories(ctx context.Context) ([]string, error) {
	return uc.repo.FindCategories()
}

func (uc *heiUseCase) Create(req *master.CreateHEIRequest) (*master.HEIMaster, error) {
	if strings.TrimSpace(req.CategoryName) == "" {
		return nil, errors.New("kategori HEI wajib diisi")
	}
	if strings.TrimSpace(req.HEIName) == "" {
		return nil, errors.New("nama item HEI wajib diisi")
	}

	status := req.Status
	if status == "" {
		status = "Active"
	}

	item := &master.HEIMaster{
		HEIID:        idgen.Generate("HEI"),
		CategoryName: strings.TrimSpace(req.CategoryName),
		HEICode:      strings.TrimSpace(req.HEICode),
		HEIName:      strings.TrimSpace(req.HEIName),
		Description:  strings.TrimSpace(req.Description),
		Status:       status,
	}

	if err := uc.repo.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *heiUseCase) Update(id string, req *master.UpdateHEIRequest) (*master.HEIMaster, error) {
	item, err := uc.repo.FindByID(id)
	if err != nil || item == nil {
		return nil, errors.New("item HEI tidak ditemukan")
	}

	if req.CategoryName != "" {
		item.CategoryName = strings.TrimSpace(req.CategoryName)
	}
	if req.HEICode != "" {
		item.HEICode = strings.TrimSpace(req.HEICode)
	}
	if req.HEIName != "" {
		item.HEIName = strings.TrimSpace(req.HEIName)
	}
	if req.Description != "" {
		item.Description = strings.TrimSpace(req.Description)
	}
	if req.Status != "" {
		item.Status = req.Status
	}

	if err := uc.repo.Update(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *heiUseCase) Delete(id string) error {
	return uc.repo.Delete(id)
}

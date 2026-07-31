package issueusecase

import (
	"errors"

	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/pkg/idgen"
)

type habitUseCase struct {
	repo issue.HabitRepository
}

func NewHabitUseCase(repo issue.HabitRepository) issue.HabitUseCase {
	return &habitUseCase{repo: repo}
}

func (u *habitUseCase) GetAll(page, limit int, search string) ([]issue.Habit, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.repo.FindAll(page, limit, search)
}

func (u *habitUseCase) GetByID(id string) (*issue.Habit, error) {
	return u.repo.FindByID(id)
}

func (u *habitUseCase) Create(req *issue.CreateHabitRequest) (*issue.Habit, error) {
	h := &issue.Habit{
		HabitID:       idgen.GenerateRandom("HBT"),
		HabitCode:     req.HabitCode,
		HabitName:     req.HabitName,
		HabitCategory: req.HabitCategory,
		Description:   req.Description,
	}

	if err := u.repo.Create(h); err != nil {
		return nil, err
	}
	return h, nil
}

func (u *habitUseCase) Update(id string, req *issue.UpdateHabitRequest) (*issue.Habit, error) {
	h, err := u.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("habit tidak ditemukan")
	}

	if req.HabitName != "" {
		h.HabitName = req.HabitName
	}
	h.HabitCode = req.HabitCode
	h.HabitCategory = req.HabitCategory
	h.Description = req.Description

	if err := u.repo.Update(h); err != nil {
		return nil, err
	}
	return h, nil
}

func (u *habitUseCase) Delete(id string) error {
	return u.repo.Delete(id)
}

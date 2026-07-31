package issueusecase

import (
	"errors"

	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/pkg/idgen"
)

type equipmentUseCase struct {
	repo issue.EquipmentRepository
}

func NewEquipmentUseCase(repo issue.EquipmentRepository) issue.EquipmentUseCase {
	return &equipmentUseCase{repo: repo}
}

func (u *equipmentUseCase) GetAll(page, limit int, kawasanID, search string) ([]issue.Equipment, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.repo.FindAll(page, limit, kawasanID, search)
}

func (u *equipmentUseCase) GetByID(id string) (*issue.Equipment, error) {
	return u.repo.FindByID(id)
}

func (u *equipmentUseCase) Create(req *issue.CreateEquipmentRequest) (*issue.Equipment, error) {
	status := req.EquipmentStatus
	if status == "" {
		status = "Active"
	}

	eq := &issue.Equipment{
		EquipmentID:     idgen.GenerateRandom("EQP"),
		KawasanID:       req.KawasanID,
		EquipmentCode:   req.EquipmentCode,
		EquipmentName:   req.EquipmentName,
		EquipmentType:   req.EquipmentType,
		EquipmentStatus: status,
	}

	if err := u.repo.Create(eq); err != nil {
		return nil, err
	}
	return eq, nil
}

func (u *equipmentUseCase) Update(id string, req *issue.UpdateEquipmentRequest) (*issue.Equipment, error) {
	eq, err := u.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("equipment tidak ditemukan")
	}

	if req.KawasanID != "" {
		eq.KawasanID = req.KawasanID
	}
	if req.EquipmentName != "" {
		eq.EquipmentName = req.EquipmentName
	}
	if req.EquipmentStatus != "" {
		eq.EquipmentStatus = req.EquipmentStatus
	}
	eq.EquipmentCode = req.EquipmentCode
	eq.EquipmentType = req.EquipmentType

	if err := u.repo.Update(eq); err != nil {
		return nil, err
	}
	return eq, nil
}

func (u *equipmentUseCase) Delete(id string) error {
	return u.repo.Delete(id)
}

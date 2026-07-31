package issueusecase

import (
	"errors"

	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/pkg/idgen"
)

type infrastructureUseCase struct {
	repo issue.InfrastructureRepository
}

func NewInfrastructureUseCase(repo issue.InfrastructureRepository) issue.InfrastructureUseCase {
	return &infrastructureUseCase{repo: repo}
}

func (u *infrastructureUseCase) GetAll(page, limit int, kawasanID, search string) ([]issue.Infrastructure, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.repo.FindAll(page, limit, kawasanID, search)
}

func (u *infrastructureUseCase) GetByID(id string) (*issue.Infrastructure, error) {
	return u.repo.FindByID(id)
}

func (u *infrastructureUseCase) Create(req *issue.CreateInfrastructureRequest) (*issue.Infrastructure, error) {
	status := req.InfrastructureStatus
	if status == "" {
		status = "Active"
	}

	inf := &issue.Infrastructure{
		InfrastructureID:     idgen.GenerateRandom("INF"),
		KawasanID:            req.KawasanID,
		InfrastructureCode:   req.InfrastructureCode,
		InfrastructureName:   req.InfrastructureName,
		InfrastructureType:   req.InfrastructureType,
		InfrastructureStatus: status,
	}

	if err := u.repo.Create(inf); err != nil {
		return nil, err
	}
	return inf, nil
}

func (u *infrastructureUseCase) Update(id string, req *issue.UpdateInfrastructureRequest) (*issue.Infrastructure, error) {
	inf, err := u.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("infrastructure tidak ditemukan")
	}

	if req.KawasanID != "" {
		inf.KawasanID = req.KawasanID
	}
	if req.InfrastructureName != "" {
		inf.InfrastructureName = req.InfrastructureName
	}
	if req.InfrastructureStatus != "" {
		inf.InfrastructureStatus = req.InfrastructureStatus
	}
	inf.InfrastructureCode = req.InfrastructureCode
	inf.InfrastructureType = req.InfrastructureType

	if err := u.repo.Update(inf); err != nil {
		return nil, err
	}
	return inf, nil
}

func (u *infrastructureUseCase) Delete(id string) error {
	return u.repo.Delete(id)
}

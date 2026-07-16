package authusecase

import (
	"errors"
	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/auth"
)

type roleUC struct {
	repo auth.RoleRepository
}

func NewRoleUseCase(repo auth.RoleRepository) auth.RoleUseCase {
	return &roleUC{repo: repo}
}

func (u *roleUC) GetAll(page, limit int, search string) ([]auth.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return u.repo.FindAll(page, limit, search)
}

func (u *roleUC) GetByID(id string) (*auth.Role, error) {
	return u.repo.FindByID(id)
}

func (u *roleUC) Create(role *auth.Role) error {
	if role.RoleName == "" {
		return errors.New("role name is required")
	}
	if role.RoleID == "" {
		role.RoleID = uuid.New().String()
	}
	return u.repo.Create(role)
}

func (u *roleUC) Update(role *auth.Role) error {
	existing, err := u.repo.FindByID(role.RoleID)
	if err != nil {
		return err
	}
	existing.RoleName = role.RoleName
	existing.RoleDescription = role.RoleDescription

	return u.repo.Update(existing)
}

func (u *roleUC) Delete(id string) error {
	return u.repo.Delete(id)
}

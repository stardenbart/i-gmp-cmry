package authusecase

import (
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
)

type rolePermUseCase struct {
	repo authdomain.RolePermissionRepository
}

// NewRolePermissionUseCase creates a new RolePermissionUseCase.
func NewRolePermissionUseCase(repo authdomain.RolePermissionRepository) authdomain.RolePermissionUseCase {
	return &rolePermUseCase{repo: repo}
}

func (uc *rolePermUseCase) GetByRoleID(roleID string) ([]authdomain.RolePermission, error) {
	return uc.repo.FindByRoleID(roleID)
}

func (uc *rolePermUseCase) SetPermission(updatedByUserID string, req *authdomain.BulkSetPermissionRequest) error {
	var rps []authdomain.RolePermission
	for _, p := range req.Permissions {
		rps = append(rps, authdomain.RolePermission{
			RoleID:                  req.RoleID,
			PermissionID:            p.PermissionID,
			IsAllowed:               p.IsAllowed,
			RolePermissionUpdatedBy: updatedByUserID,
		})
	}
	return uc.repo.BulkUpsert(rps)
}

func (uc *rolePermUseCase) CheckPermission(roleID, moduleID, permissionCode string) (bool, error) {
	return uc.repo.HasPermission(roleID, moduleID, permissionCode)
}

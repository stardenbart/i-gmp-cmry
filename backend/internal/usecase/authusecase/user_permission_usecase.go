package authusecase

import (
	"errors"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/idgen"
)

type userPermissionUseCase struct {
	upRepo authdomain.UserPermissionRepository
}

func NewUserPermissionUseCase(upRepo authdomain.UserPermissionRepository) authdomain.UserPermissionUseCase {
	return &userPermissionUseCase{upRepo: upRepo}
}

func (uc *userPermissionUseCase) GetByUserID(userID string) ([]authdomain.UserPermission, error) {
	return uc.upRepo.FindByUserID(userID)
}

func (uc *userPermissionUseCase) SetPermissionOverrides(updatedByUserID string, req *authdomain.BulkSetUserPermissionRequest) error {
	if req.UserID == "" {
		return errors.New("user_id is required")
	}

	// For simplicity, we first clear all existing overrides for this user, then insert the new ones.
	// But actually, BulkUpsert handles inserts/updates. If we want to allow removing an override,
	// maybe it's better to clear and insert. Since we are dealing with overrides, the UI might send only explicit ones.
	
	err := uc.upRepo.DeleteByUserID(req.UserID)
	if err != nil {
		return err
	}

	if len(req.Permissions) == 0 {
		return nil
	}

	var ups []authdomain.UserPermission
	for _, p := range req.Permissions {
		ups = append(ups, authdomain.UserPermission{
			UserPermissionID:        idgen.GenerateRandom("UP"),
			UserID:                  req.UserID,
			PermissionID:            p.PermissionID,
			IsAllowed:               p.IsAllowed,
			UserPermissionUpdatedBy: updatedByUserID,
		})
	}

	return uc.upRepo.BulkUpsert(ups)
}

func (uc *userPermissionUseCase) CheckOverride(userID, moduleID, permissionCode string) (*bool, error) {
	return uc.upRepo.CheckOverride(userID, moduleID, permissionCode)
}

func (uc *userPermissionUseCase) ClearOverrides(userID string) error {
	return uc.upRepo.DeleteByUserID(userID)
}

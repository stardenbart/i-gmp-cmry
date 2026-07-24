package authrepo

import (
	"database/sql"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

type userPermRepository struct{ db *gorm.DB }

func NewUserPermissionRepository(db *gorm.DB) authdomain.UserPermissionRepository {
	_ = db.AutoMigrate(&authdomain.UserPermission{})
	return &userPermRepository{db: db}
}

func (r *userPermRepository) FindByUserID(userID string) ([]authdomain.UserPermission, error) {
	var ups []authdomain.UserPermission
	err := r.db.Preload("Permission.Module").Where("\"UserID\" = ?", userID).Find(&ups).Error
	return ups, err
}

func (r *userPermRepository) CheckOverride(userID, moduleID, permissionCode string) (*bool, error) {
	var isAllowed sql.NullBool
	err := r.db.Table("\"User_Permission\" up").
		Select("up.\"IsAllowed\"").
		Joins("JOIN \"Permission_Master\" pm ON pm.\"PermissionID\" = up.\"PermissionID\"").
		Where("up.\"UserID\" = ? AND pm.\"ModuleID\" = ? AND pm.\"PermissionCode\" = ?", userID, moduleID, permissionCode).
		Scan(&isAllowed).Error

	if err != nil {
		return nil, err
	}

	if !isAllowed.Valid {
		return nil, nil // No override found
	}

	val := isAllowed.Bool
	return &val, nil
}

func (r *userPermRepository) BulkUpsert(ups []authdomain.UserPermission) error {
	return r.db.CreateInBatches(&ups, 100).Error
}

func (r *userPermRepository) Delete(id string) error {
	return r.db.Where("\"UserPermissionID\" = ?", id).Delete(&authdomain.UserPermission{}).Error
}

func (r *userPermRepository) DeleteByUserID(userID string) error {
	return r.db.Where("\"UserID\" = ?", userID).Delete(&authdomain.UserPermission{}).Error
}

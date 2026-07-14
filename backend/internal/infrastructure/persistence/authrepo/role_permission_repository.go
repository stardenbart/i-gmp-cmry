package authrepo

import (
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rolePermRepository struct{ db *gorm.DB }

func NewRolePermissionRepository(db *gorm.DB) authdomain.RolePermissionRepository {
	return &rolePermRepository{db: db}
}

func (r *rolePermRepository) FindByRoleID(roleID string) ([]authdomain.RolePermission, error) {
	var rps []authdomain.RolePermission
	err := r.db.Preload("Permission.Module").Where("RoleID = ?", roleID).Find(&rps).Error
	return rps, err
}

func (r *rolePermRepository) FindByRoleAndPermission(roleID, permissionID string) (*authdomain.RolePermission, error) {
	var rp authdomain.RolePermission
	err := r.db.Where("RoleID = ? AND PermissionID = ?", roleID, permissionID).First(&rp).Error
	return &rp, err
}

// HasPermission checks via a JOIN query whether roleID has a specific permission allowed.
func (r *rolePermRepository) HasPermission(roleID, moduleID, permissionCode string) (bool, error) {
	var count int64
	err := r.db.Table("Role_Permission rp").
		Joins("JOIN Permission_Master pm ON pm.PermissionID = rp.PermissionID").
		Where("rp.RoleID = ? AND pm.ModuleID = ? AND pm.PermissionCode = ? AND rp.IsAllowed = true", roleID, moduleID, permissionCode).
		Count(&count).Error
	return count > 0, err
}

func (r *rolePermRepository) Upsert(rp *authdomain.RolePermission) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "RoleID"}, {Name: "PermissionID"}},
		DoUpdates: clause.AssignmentColumns([]string{"IsAllowed", "RolePermissionUpdatedBy", "RolePermissionUpdatedAt"}),
	}).Create(rp).Error
}

func (r *rolePermRepository) BulkUpsert(rps []authdomain.RolePermission) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "RoleID"}, {Name: "PermissionID"}},
		DoUpdates: clause.AssignmentColumns([]string{"IsAllowed", "RolePermissionUpdatedBy", "RolePermissionUpdatedAt"}),
	}).CreateInBatches(&rps, 100).Error
}

func (r *rolePermRepository) Delete(id string) error {
	return r.db.Where("RolePermissionID = ?", id).Delete(&authdomain.RolePermission{}).Error
}

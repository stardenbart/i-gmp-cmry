package authrepo

import (
	"fmt"
	"sync"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type permCacheItem struct {
	hasPerm   bool
	expiresAt time.Time
}

var globalRolePermCache sync.Map

type rolePermRepository struct {
	db *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) authdomain.RolePermissionRepository {
	_ = db.AutoMigrate(&authdomain.RolePermission{})
	return &rolePermRepository{db: db}
}

func (r *rolePermRepository) FindByRoleID(roleID string) ([]authdomain.RolePermission, error) {
	var rps []authdomain.RolePermission
	err := r.db.Preload("Permission.Module").Where("\"RoleID\" = ?", roleID).Find(&rps).Error
	return rps, err
}

func (r *rolePermRepository) FindByRoleAndPermission(roleID, permissionID string) (*authdomain.RolePermission, error) {
	var rp authdomain.RolePermission
	err := r.db.Where("\"RoleID\" = ? AND \"PermissionID\" = ?", roleID, permissionID).Take(&rp).Error
	return &rp, err
}

// HasPermission checks via a JOIN query whether roleID has a specific permission allowed.
func (r *rolePermRepository) HasPermission(roleID, moduleID, permissionCode string) (bool, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s", roleID, moduleID, permissionCode)
	if val, ok := globalRolePermCache.Load(cacheKey); ok {
		item := val.(permCacheItem)
		if time.Now().Before(item.expiresAt) {
			return item.hasPerm, nil
		}
		globalRolePermCache.Delete(cacheKey)
	}

	var count int64
	err := r.db.Table("\"Role_Permission\" rp").
		Joins("JOIN \"Permission_Master\" pm ON pm.\"PermissionID\" = rp.\"PermissionID\"").
		Where("rp.\"RoleID\" = ? AND pm.\"ModuleID\" = ? AND pm.\"PermissionCode\" = ? AND rp.\"IsAllowed\" = true", roleID, moduleID, permissionCode).
		Count(&count).Error

	hasPerm := count > 0
	if err == nil {
		globalRolePermCache.Store(cacheKey, permCacheItem{
			hasPerm:   hasPerm,
			expiresAt: time.Now().Add(5 * time.Minute),
		})
	}
	return hasPerm, err
}

func (r *rolePermRepository) Upsert(rp *authdomain.RolePermission) error {
	globalRolePermCache = sync.Map{} // clear cache on update
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "RoleID"}, {Name: "PermissionID"}},
		DoUpdates: clause.AssignmentColumns([]string{"IsAllowed", "RolePermissionUpdatedBy", "RolePermissionUpdatedAt"}),
	}).Create(rp).Error
}

func (r *rolePermRepository) BulkUpsert(rps []authdomain.RolePermission) error {
	globalRolePermCache = sync.Map{} // clear cache on update
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "RoleID"}, {Name: "PermissionID"}},
		DoUpdates: clause.AssignmentColumns([]string{"IsAllowed", "RolePermissionUpdatedBy", "RolePermissionUpdatedAt"}),
	}).CreateInBatches(&rps, 100).Error
}

func (r *rolePermRepository) Delete(id string) error {
	globalRolePermCache = sync.Map{} // clear cache on delete
	return r.db.Where("\"RolePermissionID\" = ?", id).Delete(&authdomain.RolePermission{}).Error
}

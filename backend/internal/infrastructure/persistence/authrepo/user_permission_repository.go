package authrepo

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

type userPermCacheItem struct {
	val       *bool
	expiresAt time.Time
}

type userPermListCacheItem struct {
	ups       []authdomain.UserPermission
	expiresAt time.Time
}

var globalUserPermCache sync.Map
var globalUserPermListCache sync.Map

type userPermRepository struct{ db *gorm.DB }

func NewUserPermissionRepository(db *gorm.DB) authdomain.UserPermissionRepository {
	return &userPermRepository{db: db}
}

func (r *userPermRepository) FindByUserID(userID string) ([]authdomain.UserPermission, error) {
	if val, ok := globalUserPermListCache.Load(userID); ok {
		item := val.(userPermListCacheItem)
		if time.Now().Before(item.expiresAt) {
			return item.ups, nil
		}
		globalUserPermListCache.Delete(userID)
	}

	var ups []authdomain.UserPermission
	err := r.db.Preload("Permission.Module").Where("\"UserID\" = ?", userID).Find(&ups).Error
	if err == nil {
		globalUserPermListCache.Store(userID, userPermListCacheItem{
			ups:       ups,
			expiresAt: time.Now().Add(5 * time.Minute),
		})
	}
	return ups, err
}

func (r *userPermRepository) CheckOverride(userID, moduleID, permissionCode string) (*bool, error) {
	cacheKey := fmt.Sprintf("uperm:%s:%s:%s", userID, moduleID, permissionCode)
	if val, ok := globalUserPermCache.Load(cacheKey); ok {
		item := val.(userPermCacheItem)
		if time.Now().Before(item.expiresAt) {
			return item.val, nil
		}
		globalUserPermCache.Delete(cacheKey)
	}

	var isAllowed sql.NullBool
	err := r.db.Table("\"User_Permission\" up").
		Select("up.\"IsAllowed\"").
		Joins("JOIN \"Permission_Master\" pm ON pm.\"PermissionID\" = up.\"PermissionID\"").
		Where("up.\"UserID\" = ? AND pm.\"ModuleID\" = ? AND pm.\"PermissionCode\" = ?", userID, moduleID, permissionCode).
		Scan(&isAllowed).Error

	if err != nil {
		return nil, err
	}

	var result *bool
	if isAllowed.Valid {
		b := isAllowed.Bool
		result = &b
	}

	globalUserPermCache.Store(cacheKey, userPermCacheItem{
		val:       result,
		expiresAt: time.Now().Add(5 * time.Minute),
	})

	return result, nil
}

func (r *userPermRepository) BulkUpsert(ups []authdomain.UserPermission) error {
	clearSyncMap(&globalUserPermCache)     // clear cache on update (thread-safe)
	clearSyncMap(&globalUserPermListCache) // clear cache on update (thread-safe)
	return r.db.CreateInBatches(&ups, 100).Error
}

func (r *userPermRepository) Delete(id string) error {
	clearSyncMap(&globalUserPermCache)     // clear cache on delete (thread-safe)
	clearSyncMap(&globalUserPermListCache) // clear cache on delete (thread-safe)
	return r.db.Where("\"UserPermissionID\" = ?", id).Delete(&authdomain.UserPermission{}).Error
}

func (r *userPermRepository) DeleteByUserID(userID string) error {
	clearSyncMap(&globalUserPermCache)     // clear cache on delete (thread-safe)
	clearSyncMap(&globalUserPermListCache) // clear cache on delete (thread-safe)
	return r.db.Where("\"UserID\" = ?", userID).Delete(&authdomain.UserPermission{}).Error
}

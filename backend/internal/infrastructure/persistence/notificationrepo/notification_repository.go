package notificationrepo

import (
	"fmt"
	"sync"
	"time"

	"github.com/monitoring-system/backend/internal/domain/notification"
	"gorm.io/gorm"
)

type notifCountCacheItem struct {
	total     int64
	expiresAt time.Time
}

var globalNotifCountCache sync.Map

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) notification.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(n *notification.Notification) error {
	globalNotifCountCache.Delete(n.UserID) // invalidate cache on new notification
	return r.db.Create(n).Error
}

func (r *notificationRepository) FindByUserID(userID string, offset int, limit int) ([]notification.Notification, int64, error) {
	var items []notification.Notification
	var total int64

	cacheKey := fmt.Sprintf("notif_count:%s", userID)
	if val, ok := globalNotifCountCache.Load(cacheKey); ok {
		item := val.(notifCountCacheItem)
		if time.Now().Before(item.expiresAt) {
			total = item.total
		} else {
			globalNotifCountCache.Delete(cacheKey)
		}
	}

	query := r.db.Model(&notification.Notification{}).Where("\"UserID\" = ?", userID)

	if total == 0 {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
		globalNotifCountCache.Store(cacheKey, notifCountCacheItem{
			total:     total,
			expiresAt: time.Now().Add(10 * time.Second),
		})
	}

	err := query.Order("\"CreatedAt\" DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *notificationRepository) MarkAsRead(notificationID string) error {
	globalNotifCountCache = sync.Map{} // clear cache on update
	return r.db.Model(&notification.Notification{}).
		Where("\"NotificationID\" = ?", notificationID).
		Update("\"IsRead\"", true).Error
}

func (r *notificationRepository) MarkAllAsRead(userID string) error {
	globalNotifCountCache.Delete(fmt.Sprintf("notif_count:%s", userID))
	return r.db.Model(&notification.Notification{}).
		Where("\"UserID\" = ? AND \"IsRead\" = ?", userID, false).
		Update("\"IsRead\"", true).Error
}

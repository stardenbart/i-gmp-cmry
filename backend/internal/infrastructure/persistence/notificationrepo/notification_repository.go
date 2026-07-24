package notificationrepo

import (
	"github.com/monitoring-system/backend/internal/domain/notification"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) notification.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(n *notification.Notification) error {
	return r.db.Create(n).Error
}

func (r *notificationRepository) FindByUserID(userID string, offset int, limit int) ([]notification.Notification, int64, error) {
	var items []notification.Notification
	var total int64

	query := r.db.Model(&notification.Notification{}).Where("\"UserID\" = ?", userID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("\"CreatedAt\" DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *notificationRepository) MarkAsRead(notificationID string) error {
	return r.db.Model(&notification.Notification{}).
		Where("\"NotificationID\" = ?", notificationID).
		Update("\"IsRead\"", true).Error
}

func (r *notificationRepository) MarkAllAsRead(userID string) error {
	return r.db.Model(&notification.Notification{}).
		Where("\"UserID\" = ? AND \"IsRead\" = ?", userID, false).
		Update("\"IsRead\"", true).Error
}

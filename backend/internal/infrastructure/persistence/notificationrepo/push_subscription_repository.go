package notificationrepo

import (
	"github.com/monitoring-system/backend/internal/domain/notification"
	"gorm.io/gorm"
)

type pushSubscriptionRepository struct {
	db *gorm.DB
}

func NewPushSubscriptionRepository(db *gorm.DB) notification.PushSubscriptionRepository {
	_ = db.AutoMigrate(&notification.PushSubscription{})
	return &pushSubscriptionRepository{db: db}
}

// Save upserts by endpoint: re-subscribing the same browser/device (e.g.
// after its push keys rotate) replaces the old row instead of duplicating it.
func (r *pushSubscriptionRepository) Save(sub *notification.PushSubscription) error {
	var existing notification.PushSubscription
	err := r.db.Where(`"Endpoint" = ?`, sub.Endpoint).First(&existing).Error
	if err == nil {
		return r.db.Model(&notification.PushSubscription{}).
			Where(`"Endpoint" = ?`, sub.Endpoint).
			Updates(map[string]interface{}{
				"UserID": sub.UserID,
				"P256dh": sub.P256dh,
				"Auth":   sub.Auth,
			}).Error
	}
	return r.db.Create(sub).Error
}

func (r *pushSubscriptionRepository) FindByUserID(userID string) ([]notification.PushSubscription, error) {
	var items []notification.PushSubscription
	err := r.db.Where(`"UserID" = ?`, userID).Find(&items).Error
	return items, err
}

func (r *pushSubscriptionRepository) DeleteByEndpoint(endpoint string) error {
	return r.db.Where(`"Endpoint" = ?`, endpoint).Delete(&notification.PushSubscription{}).Error
}

package notification

import (
	"time"
)

type Notification struct {
	NotificationID string    `gorm:"column:NotificationID;primaryKey" json:"id"`
	UserID         string    `gorm:"column:UserID;not null" json:"user_id"`
	Type           string    `gorm:"column:Type;not null" json:"type"` // info, warning, error, success
	Title          string    `gorm:"column:Title;not null" json:"title"`
	Message        string    `gorm:"column:Message;not null" json:"message"`
	IsRead         bool      `gorm:"column:IsRead;default:false" json:"is_read"`
	Link           string    `gorm:"column:Link" json:"link,omitempty"`
	CreatedAt      time.Time `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
}

// TableName overrides the table name used by Notification to `Notification`
func (Notification) TableName() string {
	return "Notification"
}

// PushSubscription stores one browser/device's Web Push endpoint for a user,
// as returned by `PushManager.subscribe()`. A user can have several (one per
// device/browser they enabled notifications on).
type PushSubscription struct {
	SubscriptionID string    `gorm:"column:SubscriptionID;primaryKey" json:"id"`
	UserID         string    `gorm:"column:UserID;not null;index" json:"user_id"`
	Endpoint       string    `gorm:"column:Endpoint;not null;size:1024;uniqueIndex" json:"endpoint"`
	P256dh         string    `gorm:"column:P256dh;not null" json:"p256dh"`
	Auth           string    `gorm:"column:Auth;not null" json:"auth"`
	CreatedAt      time.Time `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
}

func (PushSubscription) TableName() string { return "Push_Subscription" }

// PushSubscriptionRequest mirrors the browser PushSubscription.toJSON() shape.
type PushSubscriptionRequest struct {
	Endpoint string `json:"endpoint" validate:"required"`
	Keys     struct {
		P256dh string `json:"p256dh" validate:"required"`
		Auth   string `json:"auth" validate:"required"`
	} `json:"keys" validate:"required"`
}

// Interfaces for Dependency Injection
type NotificationRepository interface {
	Create(notification *Notification) error
	FindByUserID(userID string, offset int, limit int) ([]Notification, int64, error)
	MarkAsRead(notificationID string) error
	MarkAllAsRead(userID string) error
}

type PushSubscriptionRepository interface {
	Save(sub *PushSubscription) error
	FindByUserID(userID string) ([]PushSubscription, error)
	DeleteByEndpoint(endpoint string) error
}

type NotificationUseCase interface {
	GetNotifications(userID string, page int, limit int) ([]Notification, int64, error)
	MarkAsRead(userID string, notificationID string) error
	MarkAllAsRead(userID string) error
	CreateSystemNotification(userID string, nType string, title string, message string, link string) error
	Subscribe(userID string, req *PushSubscriptionRequest) error
	Unsubscribe(endpoint string) error
}

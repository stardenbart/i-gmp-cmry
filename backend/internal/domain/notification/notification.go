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

// Interfaces for Dependency Injection
type NotificationRepository interface {
	Create(notification *Notification) error
	FindByUserID(userID string, offset int, limit int) ([]Notification, int64, error)
	MarkAsRead(notificationID string) error
	MarkAllAsRead(userID string) error
}

type NotificationUseCase interface {
	GetNotifications(userID string, page int, limit int) ([]Notification, int64, error)
	MarkAsRead(userID string, notificationID string) error
	MarkAllAsRead(userID string) error
	CreateSystemNotification(userID string, nType string, title string, message string, link string) error
}

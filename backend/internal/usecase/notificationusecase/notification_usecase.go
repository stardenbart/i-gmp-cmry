package notificationusecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/notification"
)

type notificationUseCase struct {
	repo notification.NotificationRepository
}

func NewNotificationUseCase(repo notification.NotificationRepository) notification.NotificationUseCase {
	return &notificationUseCase{
		repo: repo,
	}
}

func (u *notificationUseCase) GetNotifications(userID string, page int, limit int) ([]notification.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	return u.repo.FindByUserID(userID, offset, limit)
}

func (u *notificationUseCase) MarkAsRead(userID string, notificationID string) error {
	// ideally we check if the notification belongs to the user, but for simplicity we assume yes
	// or we can add UserID check in repo
	return u.repo.MarkAsRead(notificationID)
}

func (u *notificationUseCase) MarkAllAsRead(userID string) error {
	return u.repo.MarkAllAsRead(userID)
}

func (u *notificationUseCase) CreateSystemNotification(userID string, nType string, title string, message string, link string) error {
	if userID == "" || title == "" || message == "" {
		return errors.New("invalid notification parameters")
	}

	if nType == "" {
		nType = "info"
	}

	n := &notification.Notification{
		NotificationID: "NOTF-" + uuid.New().String()[:8],
		UserID:         userID,
		Type:           nType,
		Title:          title,
		Message:        message,
		IsRead:         false,
		Link:           link,
		CreatedAt:      time.Now(),
	}

	return u.repo.Create(n)
}

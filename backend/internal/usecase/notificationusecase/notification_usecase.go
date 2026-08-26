package notificationusecase

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/notification"
	"github.com/monitoring-system/backend/pkg/webpush"
)

type notificationUseCase struct {
	repo     notification.NotificationRepository
	subRepo  notification.PushSubscriptionRepository
	pushSend webpush.Sender
}

// NewNotificationUseCase wires the in-app bell (repo) together with Web Push
// delivery (subRepo + pushSend). pushSend/subRepo may be nil — in that case
// CreateSystemNotification still creates the in-app row, it just skips push.
func NewNotificationUseCase(repo notification.NotificationRepository, subRepo notification.PushSubscriptionRepository, pushSend webpush.Sender) notification.NotificationUseCase {
	return &notificationUseCase{
		repo:     repo,
		subRepo:  subRepo,
		pushSend: pushSend,
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

	if err := u.repo.Create(n); err != nil {
		return err
	}

	// Best-effort: also push to every device this user subscribed for Web
	// Push, so it shows up on their lock screen even with the app closed.
	// A failure here must never fail the notification itself — the in-app
	// bell entry above already exists regardless.
	u.pushToSubscribers(userID, title, message, link)
	return nil
}

func (u *notificationUseCase) pushToSubscribers(userID, title, message, link string) {
	if u.pushSend == nil || u.subRepo == nil {
		return
	}
	subs, err := u.subRepo.FindByUserID(userID)
	if err != nil || len(subs) == 0 {
		return
	}
	for _, sub := range subs {
		err := u.pushSend.Send(sub.Endpoint, sub.P256dh, sub.Auth, webpush.Payload{
			Title: title,
			Body:  message,
			Link:  link,
		})
		if err == webpush.ErrSubscriptionExpired {
			_ = u.subRepo.DeleteByEndpoint(sub.Endpoint)
		} else if err != nil {
			log.Printf("[Notification] push send failed for user %s: %v", userID, err)
		}
	}
}

func (u *notificationUseCase) Subscribe(userID string, req *notification.PushSubscriptionRequest) error {
	if u.subRepo == nil {
		return errors.New("push notifications are not configured on this server")
	}
	if userID == "" || req == nil || req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		return errors.New("invalid push subscription")
	}
	return u.subRepo.Save(&notification.PushSubscription{
		SubscriptionID: "PSUB-" + uuid.New().String()[:8],
		UserID:         userID,
		Endpoint:       req.Endpoint,
		P256dh:         req.Keys.P256dh,
		Auth:           req.Keys.Auth,
		CreatedAt:      time.Now(),
	})
}

func (u *notificationUseCase) Unsubscribe(endpoint string) error {
	if u.subRepo == nil {
		return errors.New("push notifications are not configured on this server")
	}
	if endpoint == "" {
		return errors.New("endpoint is required")
	}
	return u.subRepo.DeleteByEndpoint(endpoint)
}

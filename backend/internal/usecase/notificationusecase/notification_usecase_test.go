package notificationusecase

import (
	"errors"
	"testing"

	"github.com/monitoring-system/backend/internal/domain/notification"
)

// fakeNotificationRepo tracks exactly what MarkAsRead was called with, so
// the test below can assert the usecase forwards the caller's own userID —
// the fix for a broken-access-control bug where any authenticated user
// could mark another user's notification as read by ID alone.
type fakeNotificationRepo struct {
	markAsReadUserID  string
	markAsReadNotifID string
	markAsReadErr     error
}

func (r *fakeNotificationRepo) Create(*notification.Notification) error { return nil }
func (r *fakeNotificationRepo) FindByUserID(string, int, int) ([]notification.Notification, int64, error) {
	return nil, 0, nil
}
func (r *fakeNotificationRepo) MarkAsRead(userID, notificationID string) error {
	r.markAsReadUserID = userID
	r.markAsReadNotifID = notificationID
	return r.markAsReadErr
}
func (r *fakeNotificationRepo) MarkAllAsRead(string) error { return nil }

func TestMarkAsReadForwardsCallerUserID(t *testing.T) {
	repo := &fakeNotificationRepo{}
	uc := NewNotificationUseCase(repo, nil, nil)

	if err := uc.MarkAsRead("USR-1", "NOTIF-1"); err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}
	if repo.markAsReadUserID != "USR-1" {
		t.Fatalf("expected the caller's own userID (USR-1) to scope the update, got %q — without this, any user could mark ANY other user's notification as read", repo.markAsReadUserID)
	}
	if repo.markAsReadNotifID != "NOTIF-1" {
		t.Fatalf("expected notification id NOTIF-1, got %q", repo.markAsReadNotifID)
	}
}

func TestMarkAsReadPropagatesNotFoundForOtherUsersNotification(t *testing.T) {
	repo := &fakeNotificationRepo{markAsReadErr: errors.New("record not found")}
	uc := NewNotificationUseCase(repo, nil, nil)

	if err := uc.MarkAsRead("USR-1", "NOTIF-BELONGS-TO-USR-2"); err == nil {
		t.Fatal("expected an error when the notification doesn't belong to the caller")
	}
}

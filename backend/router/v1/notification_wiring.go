package v1

import (
	"github.com/monitoring-system/backend/config"
	notificationdomain "github.com/monitoring-system/backend/internal/domain/notification"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/notificationrepo"
	"github.com/monitoring-system/backend/internal/usecase/notificationusecase"
	"github.com/monitoring-system/backend/pkg/webpush"
	"gorm.io/gorm"
)

// buildNotificationUseCase wires the in-app bell repository together with Web
// Push delivery. It's shared by every route group that needs to raise a
// notification (issues, inspections, the /notifications endpoints
// themselves) so all of them push to the same subscriptions consistently.
// Without VAPID keys configured, push sending is simply skipped — the in-app
// bell still works.
func buildNotificationUseCase(db *gorm.DB, cfg *config.Config) notificationdomain.NotificationUseCase {
	repo := notificationrepo.NewNotificationRepository(db)
	subRepo := notificationrepo.NewPushSubscriptionRepository(db)

	pushEnabled := cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != ""
	pushSender := webpush.NewSender(webpush.Config{
		Enabled:    pushEnabled,
		PublicKey:  cfg.VAPIDPublicKey,
		PrivateKey: cfg.VAPIDPrivateKey,
		Subject:    cfg.VAPIDSubject,
	})

	return notificationusecase.NewNotificationUseCase(repo, subRepo, pushSender)
}

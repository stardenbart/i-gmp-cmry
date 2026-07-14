package loggingusecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/events"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/pkg/kafka"
)

// ── LoginLog UseCase ──────────────────────────────────────────────────────

type loginLogUseCase struct{ repo logdomain.LoginLogRepository }

func NewLoginLogUseCase(repo logdomain.LoginLogRepository) logdomain.LoginLogUseCase {
	return &loginLogUseCase{repo: repo}
}

func (uc *loginLogUseCase) GetAll(page, limit int, userID string) ([]logdomain.LoginLog, int64, error) {
	return uc.repo.FindAll(page, limit, userID)
}

func (uc *loginLogUseCase) GetByID(id string) (*logdomain.LoginLog, error) {
	return uc.repo.FindByID(id)
}

// ── ActivityLog UseCase (Queue-Based) ─────────────────────────────────────

type activityLogUseCase struct {
	repo     logdomain.ActivityLogRepository
	producer kafka.EventProducer
}

func NewActivityLogUseCase(repo logdomain.ActivityLogRepository, producer kafka.EventProducer) logdomain.ActivityLogUseCase {
	return &activityLogUseCase{repo: repo, producer: producer}
}

func (uc *activityLogUseCase) GetAll(page, limit int, userID, moduleID, action string) ([]logdomain.ActivityLog, int64, error) {
	return uc.repo.FindAll(page, limit, userID, moduleID, action)
}

func (uc *activityLogUseCase) GetByID(id string) (*logdomain.ActivityLog, error) {
	return uc.repo.FindByID(id)
}

func (uc *activityLogUseCase) Record(ctx context.Context, req *logdomain.CreateActivityLogRequest) error {
	event := events.ActivityLogEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeCreated,
			Timestamp: time.Now(),
			ActorID:   req.UserID,
		},
		Action:    req.ActivityAction,
		Entity:    req.TableAffected,
		EntityID:  req.RecordID,
		Details:   req.ActivityDescription,
		IPAddress: req.IPAddress,
		// Map other fields as needed if we expand the Kafka event
	}

	return uc.producer.PublishEvent(ctx, events.TopicActivityLogs, req.UserID, event)
}

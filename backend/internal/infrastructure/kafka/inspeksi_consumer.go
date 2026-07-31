package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/usecase/lockusecase"
	pkgkafka "github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/realtime"
	redis "github.com/redis/go-redis/v9"
	segmentio "github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type InspeksiConsumer struct {
	consumer *pkgkafka.EventConsumer
	lockMgr  *lockusecase.LockManager
	db       *gorm.DB
	redis    *redis.Client
	log      *logger.Logger
}

func NewInspeksiConsumer(brokers []string, groupID string, lockMgr *lockusecase.LockManager, db *gorm.DB, r *redis.Client, log *logger.Logger) *InspeksiConsumer {
	consumer := pkgkafka.NewConsumer(brokers, groupID, events.TopicInspeksiAspekSave, log)
	return &InspeksiConsumer{
		consumer: consumer,
		lockMgr:  lockMgr,
		db:       db,
		redis:    r,
		log:      log,
	}
}

func (c *InspeksiConsumer) Start(ctx context.Context) {
	c.consumer.Start(ctx, c.handleMessage)
}

func (c *InspeksiConsumer) handleMessage(ctx context.Context, msg segmentio.Message) error {
	var req inspection.SaveAspekRequest
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		c.log.Error("Failed to parse InspeksiMessage", logger.Error(err))
		return nil // Non-retryable parse error
	}

	// 1. Idempotency Check
	if c.redis != nil && req.LockToken != "" {
		msgID := string(msg.Key) + ":" + fmt.Sprintf("%d", msg.Time.UnixNano())
		idemKey := "idem:" + msgID
		set, err := c.redis.SetNX(ctx, idemKey, "1", time.Hour).Result()
		if err == nil && !set {
			c.log.Info("Duplicate message skipped", logger.String("msg_id", msgID))
			return nil
		}
	}

	// 2. Validate Lock
	if err := c.lockMgr.ValidateLock(ctx, req.KawasanID, req.AspekID, req.UserID, req.LockToken); err != nil {
		c.log.Warn("Lock invalid or expired during message processing", logger.String("aspek_id", req.AspekID), logger.Error(err))
		return nil
	}

	// 3. Save draft state to Redis
	dataBytes, _ := json.Marshal(req.Data)
	if err := c.lockMgr.SaveDraftState(ctx, req.KawasanID, req.AspekID, req.UserID, string(dataBytes), req.Skor); err != nil {
		c.log.Error("Failed to save draft state in Redis", logger.Error(err))
		return err
	}

	// Notify User via SSE that draft was saved
	realtime.GetSSEBroker().SendToUser(req.UserID, realtime.SSEEvent{
		Type:    "SAVE_SUCCESS",
		AspekID: req.AspekID,
		Message: "Draft aspek berhasil disimpan",
	})

	// 4. Handle Completion
	if req.IsFinal {
		periodeStr := time.Now().Format("2006-01-01")
		allDone, err := c.lockMgr.MarkAspekCompleted(ctx, req.KawasanID, req.AspekID, periodeStr)
		if err != nil {
			c.log.Error("Failed to mark aspek completed", logger.Error(err))
			return err
		}

		// Broadcast Aspek Completed via WebSocket
		realtime.GetWSHub().BroadcastToKawasan(req.KawasanID, realtime.WSEvent{
			Type:      "ASPEK_COMPLETED",
			KawasanID: req.KawasanID,
			AspekID:   req.AspekID,
			UserID:    req.UserID,
		})

		// If all aspek in Kawasan are done, sync to DB Core!
		if allDone {
			c.log.Info("All aspek completed for Kawasan! Triggering DB Core Sync...", logger.String("kawasan_id", req.KawasanID))
			go c.SyncKawasanToDB(context.Background(), req.KawasanID, req.SessionID, periodeStr)
		}
	}

	return nil
}

// SyncKawasanToDB flushes Redis draft states to PostgreSQL in a single serializable transaction
func (c *InspeksiConsumer) SyncKawasanToDB(ctx context.Context, kawasanID, sessionID, periode string) error {
	syncLockKey := fmt.Sprintf("sync:%s:%s", kawasanID, periode)

	if c.redis != nil {
		ok, err := c.redis.SetNX(ctx, syncLockKey, sessionID, 5*time.Minute).Result()
		if err != nil || !ok {
			c.log.Info("Sync already in progress for Kawasan", logger.String("kawasan_id", kawasanID))
			return nil
		}
		defer c.redis.Del(ctx, syncLockKey)
	}

	// Read tracker keys
	trackerKey := fmt.Sprintf("tracker:%s:%s", kawasanID, periode)
	var aspekIDs []string
	if c.redis != nil {
		aspekMap, err := c.redis.HGetAll(ctx, trackerKey).Result()
		if err == nil {
			for k := range aspekMap {
				aspekIDs = append(aspekIDs, k)
			}
		}
	}

	// Fallback to fetch master aspek for Kawasan if Redis tracker is empty
	if len(aspekIDs) == 0 {
		var kawasanAspeks []inspection.KawasanAspek
		c.db.Where("KawasanID = ?", kawasanID).Find(&kawasanAspeks)
		for _, ka := range kawasanAspeks {
			aspekIDs = append(aspekIDs, ka.AspekID)
		}
	}

	var results []inspection.InspeksiAspekResult
	for _, aspekID := range aspekIDs {
		draftMap, err := c.lockMgr.GetDraftState(ctx, kawasanID, aspekID)
		if err != nil || len(draftMap) == 0 {
			continue
		}

		resID := uuid.New().String()
		results = append(results, inspection.InspeksiAspekResult{
			ResultID:     resID,
			SessionID:    sessionID,
			KawasanID:    kawasanID,
			AspekID:      aspekID,
			UserID:       draftMap["user_id"],
			DataInspeksi: draftMap["data"],
			CompletedAt:  time.Now(),
		})
	}

	// Begin DB Transaction
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// 1. Update Session status
		if err := tx.Model(&inspection.InspeksiSession{}).
			Where("SessionID = ?", sessionID).
			Updates(map[string]interface{}{
				"Status":   "Synced",
				"SyncedAt": time.Now(),
			}).Error; err != nil {
			return err
		}

		// 2. Insert Results
		for _, r := range results {
			if err := tx.Create(&r).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.log.Error("DB Sync failed", logger.Error(err))
		return err
	}

	c.log.Info("DB Sync completed successfully for Kawasan", logger.String("kawasan_id", kawasanID))

	// WebSocket broadcast: KAWASAN_SYNCED
	realtime.GetWSHub().BroadcastToKawasan(kawasanID, realtime.WSEvent{
		Type:      "KAWASAN_SYNCED",
		KawasanID: kawasanID,
	})

	// SSE broadcast to users
	for _, r := range results {
		realtime.GetSSEBroker().SendToUser(r.UserID, realtime.SSEEvent{
			Type:    "KAWASAN_DONE",
			Message: fmt.Sprintf("Inspeksi kawasan %s telah selesai & tersimpan di database", kawasanID),
		})
	}

	// Cleanup Redis state keys
	if c.redis != nil {
		go func() {
			c.redis.Del(ctx, trackerKey)
			c.redis.Del(ctx, fmt.Sprintf("session:%s:%s", kawasanID, periode))
			for _, aID := range aspekIDs {
				c.redis.Del(ctx, fmt.Sprintf("state:aspek:%s:%s", kawasanID, aID))
				c.redis.Del(ctx, fmt.Sprintf("lock:aspek:%s:%s", kawasanID, aID))
			}
		}()
	}

	return nil
}

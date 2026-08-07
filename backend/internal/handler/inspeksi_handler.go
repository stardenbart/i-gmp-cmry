package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/usecase/lockusecase"
	"github.com/monitoring-system/backend/pkg/eventstore"
	pkgkafka "github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type InspeksiHandler struct {
	lockMgr       *lockusecase.LockManager
	db            *gorm.DB
	kafkaProducer pkgkafka.EventProducer
	log           *logger.Logger
	rdb           *redis.Client
}

func NewInspeksiHandler(lockMgr *lockusecase.LockManager, db *gorm.DB, producer pkgkafka.EventProducer, log *logger.Logger, rdb *redis.Client) *InspeksiHandler {
	return &InspeksiHandler{
		lockMgr:       lockMgr,
		db:            db,
		kafkaProducer: producer,
		log:           log,
		rdb:           rdb,
	}
}

// AcquireLock - POST /api/v1/inspeksi/:kawasanId/:aspekId/lock
func (h *InspeksiHandler) AcquireLock(c *fiber.Ctx) error {
	kawasanID := c.Params("kawasanId")
	aspekID := c.Params("aspekId")
	userID, _ := c.Locals("userID").(string)

	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	result, err := h.lockMgr.AcquireLock(c.Context(), kawasanID, aspekID, userID)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	// Push LOCK_ACQUIRED event to Kawasan event stream
	eventstore.GetEventStore(h.rdb).PushKawasan(kawasanID, eventstore.KawasanEvent{
		Type:      "LOCK_ACQUIRED",
		KawasanID: kawasanID,
		AspekID:   aspekID,
		UserID:    userID,
		Payload:   result,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// RenewLock - PUT /api/v1/inspeksi/:kawasanId/:aspekId/lock/heartbeat
func (h *InspeksiHandler) RenewLock(c *fiber.Ctx) error {
	kawasanID := c.Params("kawasanId")
	aspekID := c.Params("aspekId")
	userID, _ := c.Locals("userID").(string)
	token := c.Get("X-Lock-Token")

	if err := h.lockMgr.RenewLock(c.Context(), kawasanID, aspekID, userID, token); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Lock renewed"})
}

// ReleaseLock - DELETE /api/v1/inspeksi/:kawasanId/:aspekId/lock
func (h *InspeksiHandler) ReleaseLock(c *fiber.Ctx) error {
	kawasanID := c.Params("kawasanId")
	aspekID := c.Params("aspekId")
	userID, _ := c.Locals("userID").(string)
	token := c.Get("X-Lock-Token")

	if err := h.lockMgr.ReleaseLock(c.Context(), kawasanID, aspekID, userID, token); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	// Push LOCK_RELEASED event to Kawasan event stream
	eventstore.GetEventStore(h.rdb).PushKawasan(kawasanID, eventstore.KawasanEvent{
		Type:      "LOCK_RELEASED",
		KawasanID: kawasanID,
		AspekID:   aspekID,
		UserID:    userID,
	})

	return c.JSON(fiber.Map{"success": true, "message": "Lock released"})
}

// YieldRequest - POST /api/v1/inspeksi/:kawasanId/:aspekId/yield-request
func (h *InspeksiHandler) YieldRequest(c *fiber.Ctx) error {
	kawasanID := c.Params("kawasanId")
	aspekID := c.Params("aspekId")
	requesterID, _ := c.Locals("userID").(string)

	lockInfo, err := h.lockMgr.GetLockInfo(c.Context(), kawasanID, aspekID)
	if err != nil || lockInfo == nil || lockInfo.Status != "LOCKED" {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"message": "Aspek sudah tidak dikunci"})
	}

	// Push YIELD_REQUEST to lock holder's event stream
	eventstore.GetEventStore(h.rdb).PushUser(lockInfo.LockedBy, eventstore.UserEvent{
		Type:    "YIELD_REQUEST",
		AspekID: aspekID,
		Message: "User lain meminta giliran untuk mengisi aspek ini",
		Payload: map[string]string{
			"requested_by": requesterID,
			"kawasan_id":   kawasanID,
			"aspek_id":     aspekID,
		},
	})

	return c.JSON(fiber.Map{"success": true, "message": "Permintaan giliran terkirim"})
}

// SaveAspek - PUT /api/v1/inspeksi/:kawasanId/:aspekId
func (h *InspeksiHandler) SaveAspek(c *fiber.Ctx) error {
	kawasanID := c.Params("kawasanId")
	aspekID := c.Params("aspekId")
	userID, _ := c.Locals("userID").(string)
	lockToken := c.Get("X-Lock-Token")

	// Validate Lock
	if err := h.lockMgr.ValidateLock(c.Context(), kawasanID, aspekID, userID, lockToken); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	var req inspection.SaveAspekRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	req.KawasanID = kawasanID
	req.AspekID = aspekID
	req.UserID = userID
	req.LockToken = lockToken

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	// Produce message to Kafka
	if h.kafkaProducer != nil {
		msgBytes, _ := json.Marshal(req)
		key := kawasanID
		if err := h.kafkaProducer.PublishEvent(c.Context(), events.TopicInspeksiAspekSave, key, string(msgBytes)); err != nil {
			h.log.Warn("Kafka produce warning, saving draft state directly", logger.Error(err))
		}
	}

	// Immediate draft save to Redis for ultra-low latency response
	dataBytes, _ := json.Marshal(req.Data)
	h.lockMgr.SaveDraftState(c.Context(), kawasanID, aspekID, userID, string(dataBytes), req.Skor)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data aspek berhasil diproses",
	})
}

// GetKawasanStatus - GET /api/v1/inspeksi/:kawasanId/status
func (h *InspeksiHandler) GetKawasanStatus(c *fiber.Ctx) error {
	kawasanID := c.Params("kawasanId")

	var kawasanAspeks []inspection.KawasanAspek
	h.db.Where("KawasanID = ?", kawasanID).Find(&kawasanAspeks)

	statuses := make([]inspection.AspekLockStatus, 0, len(kawasanAspeks))
	for _, ka := range kawasanAspeks {
		info, err := h.lockMgr.GetLockInfo(c.Context(), kawasanID, ka.AspekID)
		if err != nil || info == nil {
			statuses = append(statuses, inspection.AspekLockStatus{
				AspekID: ka.AspekID,
				Status:  "FREE",
			})
		} else {
			statuses = append(statuses, *info)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    statuses,
	})
}

// GetDraftState - GET /api/v1/inspeksi/:kawasanId/:aspekId/state
func (h *InspeksiHandler) GetDraftState(c *fiber.Ctx) error {
	scopeID := c.Params("kawasanId")
	aspekID := c.Params("aspekId")

	draft, err := h.lockMgr.GetDraftState(c.Context(), scopeID, aspekID)
	if (err != nil || len(draft) == 0) && h.db != nil {
		var header inspection.InspectionHeader
		if dbErr := h.db.Where("InspectionID = ?", scopeID).First(&header).Error; dbErr == nil && header.KawasanID != "" {
			legacyDraft, _ := h.lockMgr.GetDraftState(c.Context(), header.KawasanID, aspekID)
			if len(legacyDraft) > 0 {
				draft = legacyDraft
			}
		}
	}
	if err != nil && len(draft) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	lockInfo, _ := h.lockMgr.GetLockInfo(c.Context(), scopeID, aspekID)

	return c.JSON(fiber.Map{
		"success":   true,
		"draft":     draft,
		"lock_info": lockInfo,
	})
}

// GetAllAspekDraftState - GET /api/v1/inspeksi/:kawasanId/drafts
func (h *InspeksiHandler) GetAllAspekDraftState(c *fiber.Ctx) error {
	scopeID := c.Params("kawasanId")

	drafts := make(map[string]map[string]string)
	if h.rdb != nil {
		pattern := fmt.Sprintf("state:aspek:%s:*", scopeID)
		keys, err := h.rdb.Keys(c.Context(), pattern).Result()

		// Fallback / Auto-Migration: If no keys found under scopeID (e.g. inspectionId),
		// check if scopeID corresponds to an inspection header and migrate legacy keys from kawasanID.
		if (err != nil || len(keys) == 0) && h.db != nil {
			var header inspection.InspectionHeader
			if dbErr := h.db.Where("InspectionID = ?", scopeID).First(&header).Error; dbErr == nil && header.KawasanID != "" {
				legacyPattern := fmt.Sprintf("state:aspek:%s:*", header.KawasanID)
				legacyKeys, _ := h.rdb.Keys(c.Context(), legacyPattern).Result()
				if len(legacyKeys) > 0 {
					for _, lKey := range legacyKeys {
						parts := strings.Split(lKey, ":")
						if len(parts) >= 4 {
							aspekID := parts[3]
							newKey := fmt.Sprintf("state:aspek:%s:%s", scopeID, aspekID)
							// Retrieve legacy hash and write to new scope key
							if legacyData, getErr := h.rdb.HGetAll(c.Context(), lKey).Result(); getErr == nil && len(legacyData) > 0 {
								args := make([]interface{}, 0, len(legacyData)*2)
								for k, v := range legacyData {
									args = append(args, k, v)
								}
								if setErr := h.rdb.HSet(c.Context(), newKey, args...).Err(); setErr != nil {
									h.log.Warn("Failed to copy legacy Redis draft key", logger.Error(setErr))
								} else {
									_ = h.rdb.Expire(c.Context(), newKey, 24*time.Hour)
								}
							}
						}
					}
					// Re-query keys under new scopeID
					keys, _ = h.rdb.Keys(c.Context(), pattern).Result()
				}
			}
		}

		if len(keys) > 0 {
			// Batch Pipeline to read all Redis hashes in a single round-trip chunk
			pipe := h.rdb.Pipeline()
			cmds := make(map[string]*redis.MapStringStringCmd)
			for _, key := range keys {
				parts := strings.Split(key, ":")
				if len(parts) >= 4 {
					aspekID := parts[3]
					cmds[aspekID] = pipe.HGetAll(c.Context(), key)
				}
			}
			_, _ = pipe.Exec(c.Context())

			for aspekID, cmd := range cmds {
				draft, err := cmd.Result()
				if err == nil && len(draft) > 0 {
					drafts[aspekID] = draft
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    drafts,
	})
}

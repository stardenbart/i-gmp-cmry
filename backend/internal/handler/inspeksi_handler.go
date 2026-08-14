package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
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
	"golang.org/x/sync/singleflight"
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

	// Immediate draft save to Redis DULU — ini jalur utama, harus ultra-low latency
	dataBytes, _ := json.Marshal(req.Data)
	h.lockMgr.SaveDraftState(c.Context(), kawasanID, aspekID, userID, string(dataBytes), req.Skor)

	// Publish ke Kafka secara ASYNC (fire-and-forget) agar tidak memblok HTTP response.
	// Draft sudah aman tersimpan di Redis di atas. Kafka hanya untuk event stream / audit.
	if h.kafkaProducer != nil {
		reqCopy := req // capture copy untuk goroutine
		keyCopy := kawasanID
		go func() {
			if err := h.kafkaProducer.PublishEvent(context.Background(), events.TopicInspeksiAspekSave, keyCopy, reqCopy); err != nil {
				h.log.Warn("Kafka async produce warning", logger.Error(err))
			}
		}()
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data aspek berhasil diproses",
	})
}

type kawasanAspekCacheItem struct {
	aspeks    []inspection.KawasanAspek
	expiresAt time.Time
}

var globalKawasanAspekCache sync.Map

// draftMigrationSFGroup coalesces concurrent DB lookups for the same scopeID
// saat Redis kosong — mencegah thundering herd ke DB pool dari 200 VU bersamaan.
var draftMigrationSFGroup singleflight.Group

// looksLikeInspectionID returns true jika scopeID berformat InspectionID
// (misal INSP-LT-001-01) bukan KawasanID (misal KWS-001).
// Heuristic: InspectionID mengandung ≥3 segmen dipisah '-' dan dimulai 'INSP'.
// Ini mencegah DB fallback yang sia-sia saat caller sudah menggunakan KawasanID.
func looksLikeInspectionID(s string) bool {
	return strings.HasPrefix(strings.ToUpper(s), "INSP") && strings.Count(s, "-") >= 3
}

// GetKawasanStatus - GET /api/v1/inspeksi/:kawasanId/status
func (h *InspeksiHandler) GetKawasanStatus(c *fiber.Ctx) error {
	kawasanID := c.Params("kawasanId")

	var kawasanAspeks []inspection.KawasanAspek
	if val, ok := globalKawasanAspekCache.Load(kawasanID); ok {
		item := val.(kawasanAspekCacheItem)
		if time.Now().Before(item.expiresAt) {
			kawasanAspeks = item.aspeks
		} else {
			globalKawasanAspekCache.Delete(kawasanID)
		}
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 3*time.Second)
	defer cancel()

	if len(kawasanAspeks) == 0 {
		h.db.WithContext(ctx).Where("\"KawasanID\" = ?", kawasanID).Find(&kawasanAspeks)
		if len(kawasanAspeks) > 0 {
			globalKawasanAspekCache.Store(kawasanID, kawasanAspekCacheItem{
				aspeks:    kawasanAspeks,
				expiresAt: time.Now().Add(10 * time.Minute),
			})
		}
	}

	// Ambil semua lock status dalam SATU pipeline (bukan 5 serial GetLockInfo yang masing-masing
	// melakukan 1 pipeline GET+TTL = 5 round-trips total untuk 5 aspek)
	// Dengan bulk pipeline: semua key dikirim dalam 1 round-trip
	statuses := make([]inspection.AspekLockStatus, 0, len(kawasanAspeks))
	if len(kawasanAspeks) > 0 && h.lockMgr != nil {
		pipe := h.rdb.Pipeline()
		type pipeEntry struct {
			aspekID string
			getCmd  *redis.StringCmd
			ttlCmd  *redis.DurationCmd
		}
		entries := make([]pipeEntry, 0, len(kawasanAspeks))
		for _, ka := range kawasanAspeks {
			key := fmt.Sprintf("lock:aspek:%s:%s", kawasanID, ka.AspekID)
			entries = append(entries, pipeEntry{
				aspekID: ka.AspekID,
				getCmd:  pipe.Get(ctx, key),
				ttlCmd:  pipe.TTL(ctx, key),
			})
		}
		_, _ = pipe.Exec(ctx)

		for _, e := range entries {
			val, err := e.getCmd.Result()
			if err != nil || val == "" {
				statuses = append(statuses, inspection.AspekLockStatus{
					AspekID: e.aspekID,
					Status:  "FREE",
				})
			} else {
				parts := strings.SplitN(val, ":", 3)
				ttl, _ := e.ttlCmd.Result()
				expiresAt := time.Now().Add(ttl)
				statuses = append(statuses, inspection.AspekLockStatus{
					AspekID:   e.aspekID,
					Status:    "LOCKED",
					LockedBy:  parts[0],
					ExpiresAt: &expiresAt,
				})
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    statuses,
	})
}

// GetDraftState - GET /api/v1/inspeksi/:kawasanId/:aspekId/state
func (h *InspeksiHandler) GetDraftState(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 3*time.Second)
	defer cancel()

	scopeID := c.Params("kawasanId")
	aspekID := c.Params("aspekId")

	draft, err := h.lockMgr.GetDraftState(ctx, scopeID, aspekID)
	if (err != nil || len(draft) == 0) && h.db != nil {
		var header inspection.InspectionHeader
		if dbErr := h.db.WithContext(ctx).Where("\"InspectionID\" = ?", scopeID).First(&header).Error; dbErr == nil && header.KawasanID != "" {
			legacyDraft, _ := h.lockMgr.GetDraftState(ctx, header.KawasanID, aspekID)
			if len(legacyDraft) > 0 {
				draft = legacyDraft
			}
		}
	}
	if err != nil && len(draft) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	lockInfo, _ := h.lockMgr.GetLockInfo(ctx, scopeID, aspekID)

	return c.JSON(fiber.Map{
		"success":   true,
		"draft":     draft,
		"lock_info": lockInfo,
	})
}

// GetAllAspekDraftState - GET /api/v1/inspeksi/:kawasanId/drafts
func (h *InspeksiHandler) GetAllAspekDraftState(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 3*time.Second)
	defer cancel()

	scopeID := c.Params("kawasanId")

	drafts := make(map[string]map[string]string)
	if h.rdb != nil {
		pattern := fmt.Sprintf("state:aspek:%s:*", scopeID)
		
		// Gunakan SCAN Iterator (non-blocking O(1) step) pengganti KEYS O(N) anti-pattern
		var keys []string
		iter := h.rdb.Scan(ctx, 0, pattern, 100).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		err := iter.Err()

		// Fallback / Auto-Migration: hanya dijalankan jika:
		//   1. SCAN tidak menemukan key (belum ada draft atau scope baru)
		//   2. scopeID berbentuk InspectionID (misal INSP-LT-001-01) — bukan KawasanID
		//      Kalau caller sudah pakai KawasanID (KWS-001), SCAN 0 hasil = belum ada draft → stop.
		//   3. Dikemas dalam singleflight agar 200 VU yang bersamaan hanya memicu 1 DB query
		if (err != nil || len(keys) == 0) && h.db != nil && looksLikeInspectionID(scopeID) {
			type migResult struct {
				kawasanID string
			}
			v, _, _ := draftMigrationSFGroup.Do(scopeID, func() (interface{}, error) {
				var header inspection.InspectionHeader
				if dbErr := h.db.WithContext(ctx).Where("\"InspectionID\" = ?", scopeID).First(&header).Error; dbErr == nil {
					return &migResult{kawasanID: header.KawasanID}, nil
				}
				return &migResult{}, nil
			})

			if res, ok := v.(*migResult); ok && res.kawasanID != "" {
				legacyPattern := fmt.Sprintf("state:aspek:%s:*", res.kawasanID)
				var legacyKeys []string
				legIter := h.rdb.Scan(ctx, 0, legacyPattern, 100).Iterator()
				for legIter.Next(ctx) {
					legacyKeys = append(legacyKeys, legIter.Val())
				}
				if len(legacyKeys) > 0 {
					for _, lKey := range legacyKeys {
						parts := strings.Split(lKey, ":")
						if len(parts) >= 4 {
							aspekID := parts[3]
							newKey := fmt.Sprintf("state:aspek:%s:%s", scopeID, aspekID)
							if legacyData, getErr := h.rdb.HGetAll(ctx, lKey).Result(); getErr == nil && len(legacyData) > 0 {
								args := make([]interface{}, 0, len(legacyData)*2)
								for k, v := range legacyData {
									args = append(args, k, v)
								}
								if setErr := h.rdb.HSet(ctx, newKey, args...).Err(); setErr != nil {
									h.log.Warn("Failed to copy legacy Redis draft key", logger.Error(setErr))
								} else {
									_ = h.rdb.Expire(ctx, newKey, 24*time.Hour)
								}
							}
						}
					}
					// Re-scan keys under new scopeID after migration
					keys = nil
					reIter := h.rdb.Scan(ctx, 0, pattern, 100).Iterator()
					for reIter.Next(ctx) {
						keys = append(keys, reIter.Val())
					}
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
					cmds[aspekID] = pipe.HGetAll(ctx, key)
				}
			}
			_, _ = pipe.Exec(ctx)

			for aspekID, cmd := range cmds {
				draft, err := cmd.Result()
				if err == nil && len(draft) > 0 {
					drafts[aspekID] = draft
				}
			}
		}
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"message": "service busy, please try again",
			"error":   "context deadline exceeded",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    drafts,
	})
}

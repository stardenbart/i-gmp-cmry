package handler

import (
	"bufio"
	"encoding/json"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/usecase/lockusecase"
	pkgkafka "github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/realtime"
	"gorm.io/gorm"
)

type InspeksiHandler struct {
	lockMgr       *lockusecase.LockManager
	db            *gorm.DB
	kafkaProducer pkgkafka.EventProducer
	log           *logger.Logger
}

func NewInspeksiHandler(lockMgr *lockusecase.LockManager, db *gorm.DB, producer pkgkafka.EventProducer, log *logger.Logger) *InspeksiHandler {
	return &InspeksiHandler{
		lockMgr:       lockMgr,
		db:            db,
		kafkaProducer: producer,
		log:           log,
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

	// Broadcast LOCK_ACQUIRED via WebSocket
	realtime.GetWSHub().BroadcastToKawasan(kawasanID, realtime.WSEvent{
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

	// Broadcast LOCK_RELEASED via WebSocket
	realtime.GetWSHub().BroadcastToKawasan(kawasanID, realtime.WSEvent{
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

	// Send SSE to lock holder
	realtime.GetSSEBroker().SendToUser(lockInfo.LockedBy, realtime.SSEEvent{
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
	kawasanID := c.Params("kawasanId")
	aspekID := c.Params("aspekId")

	draft, err := h.lockMgr.GetDraftState(c.Context(), kawasanID, aspekID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	lockInfo, _ := h.lockMgr.GetLockInfo(c.Context(), kawasanID, aspekID)

	return c.JSON(fiber.Map{
		"success":   true,
		"draft":     draft,
		"lock_info": lockInfo,
	})
}

// WebSocket Endpoint - GET /ws/inspeksi/:kawasanId
func (h *InspeksiHandler) WebSocketGrid(c *websocket.Conn) {
	kawasanID := c.Params("kawasanId")
	hub := realtime.GetWSHub()

	hub.Register(kawasanID, c)
	defer hub.Unregister(kawasanID, c)

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}

// SSE Endpoint - GET /api/v1/sse/inspeksi/notifications
func (h *InspeksiHandler) SSENotifications(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	if userID == "" {
		userID = c.Query("user_id")
	}
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id required"})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	broker := realtime.GetSSEBroker()
	ch := broker.Register(userID)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		stopCh := make(chan struct{})
		realtime.StreamToClient(w, ch, stopCh)
		broker.Unregister(userID, ch)
	})

	return nil
}

package pollinghandler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/eventstore"
	redis "github.com/redis/go-redis/v9"
)

type PollingHandler struct {
	store *eventstore.EventStore
}

func NewPollingHandler(rdb *redis.Client) *PollingHandler {
	return &PollingHandler{
		store: eventstore.GetEventStore(rdb),
	}
}

// PollGlobalEvents handles GET /api/v1/events/poll?since=<timestamp_ms>
func (h *PollingHandler) PollGlobalEvents(c *fiber.Ctx) error {
	sinceStr := c.Query("since")
	var sinceMs int64
	if sinceStr != "" {
		sinceMs, _ = strconv.ParseInt(sinceStr, 10, 64)
	}

	events := h.store.GetGlobalSince(sinceMs)
	now := time.Now().UnixMilli()

	return c.JSON(fiber.Map{
		"server_time": now,
		"events":      events,
	})
}

// PollKawasanEvents handles GET /api/v1/inspeksi/kawasan/:kawasanId/poll?since=<timestamp_ms>
func (h *PollingHandler) PollKawasanEvents(c *fiber.Ctx) error {
	kawasanID := c.Params("kawasanId")
	if kawasanID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "kawasanId is required",
		})
	}

	sinceStr := c.Query("since")
	var sinceMs int64
	if sinceStr != "" {
		sinceMs, _ = strconv.ParseInt(sinceStr, 10, 64)
	}

	events := h.store.GetKawasanSince(kawasanID, sinceMs)
	now := time.Now().UnixMilli()

	return c.JSON(fiber.Map{
		"server_time": now,
		"kawasan_id":  kawasanID,
		"events":      events,
	})
}

// PollUserNotifications handles GET /api/v1/inspeksi/notifications/poll?since=<timestamp_ms>
func (h *PollingHandler) PollUserNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		if u, ok := c.Locals("userID").(string); ok {
			userID = u
		}
	}
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "user_id is required",
		})
	}

	sinceStr := c.Query("since")
	var sinceMs int64
	if sinceStr != "" {
		sinceMs, _ = strconv.ParseInt(sinceStr, 10, 64)
	}

	notifications := h.store.GetUserSince(userID, sinceMs)
	now := time.Now().UnixMilli()

	return c.JSON(fiber.Map{
		"server_time":   now,
		"user_id":       userID,
		"notifications": notifications,
	})
}

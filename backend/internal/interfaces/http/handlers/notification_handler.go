package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/notification"
	"github.com/monitoring-system/backend/pkg/response"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	usecase notification.NotificationUseCase
}

func NewNotificationHandler(usecase notification.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{
		usecase: usecase,
	}
}

// GetNotifications returns notifications for the authenticated user
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	// Extract userID from locals (set by auth middleware)
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return response.Unauthorized(c, "Unauthorized")
	}
	userID := userIDVal.(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	items, total, err := h.usecase.GetNotifications(userID, page, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch notifications", err.Error())
	}

	return response.Paginated(c, "Notifications fetched successfully", items, total, page, limit)
}

// MarkAsRead marks a single notification as read
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return response.Unauthorized(c, "Unauthorized")
	}
	userID := userIDVal.(string)
	notifID := c.Params("id")

	if err := h.usecase.MarkAsRead(userID, notifID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NotFound(c, "notification not found")
		}
		return response.InternalServerError(c, "Failed to mark as read", err.Error())
	}

	return response.OK(c, "Notification marked as read", nil)
}

// MarkAllAsRead marks all notifications as read for the user
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return response.Unauthorized(c, "Unauthorized")
	}
	userID := userIDVal.(string)

	if err := h.usecase.MarkAllAsRead(userID); err != nil {
		return response.InternalServerError(c, "Failed to mark all as read", err.Error())
	}

	return response.OK(c, "All notifications marked as read", nil)
}

// Subscribe registers a browser Web Push subscription for the authenticated
// user, so future notifications also show up as OS-level push notifications.
func (h *NotificationHandler) Subscribe(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return response.Unauthorized(c, "Unauthorized")
	}
	userID := userIDVal.(string)

	var req notification.PushSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	if err := h.usecase.Subscribe(userID, &req); err != nil {
		return response.BadRequest(c, "Failed to save push subscription", err.Error())
	}

	return response.OK(c, "Push subscription saved", nil)
}

// Unsubscribe removes a Web Push subscription (e.g. the user disabled
// notifications, or the browser's subscription rotated/expired).
func (h *NotificationHandler) Unsubscribe(c *fiber.Ctx) error {
	var req struct {
		Endpoint string `json:"endpoint" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	if err := h.usecase.Unsubscribe(req.Endpoint); err != nil {
		return response.BadRequest(c, "Failed to remove push subscription", err.Error())
	}

	return response.OK(c, "Push subscription removed", nil)
}

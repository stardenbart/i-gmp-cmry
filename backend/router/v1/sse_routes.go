package v1

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/monitoring-system/backend/internal/handler/ssehandler"
	"github.com/monitoring-system/backend/pkg/sse"
)

func RegisterSSERoutes(rg fiber.Router, broker *sse.Broker) {
	h := ssehandler.NewSSEHandler(broker)

	// Rate limiting specifically for SSE endpoint to prevent connection spam
	sseLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": "Too many SSE connection attempts",
			})
		},
	})

	sseGroup := rg.Group("/sse")
	sseGroup.Use(sseLimiter)

	// We might not want the JWT middleware blocking SSE if EventSource doesn't send headers.
	// But let's assume EventSource polyfill or query tokens are used, or it's public/cookie auth.
	// For now, keep it simple. Usually SSE authentication is done via query params.
	sseGroup.Get("/events", h.StreamEvents)
}

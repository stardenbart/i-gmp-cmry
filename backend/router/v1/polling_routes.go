package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/pollinghandler"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/jwt"
	redis "github.com/redis/go-redis/v9"
)

func RegisterPollingRoutes(rg fiber.Router, rdb *redis.Client, jwtManager *jwt.Manager) {
	h := pollinghandler.NewPollingHandler(rdb)
	authMW := middleware.AuthMiddleware(jwtManager)

	// Global events poll endpoint (Issues, Dashboard, Inspections)
	rg.Get("/events/poll", h.PollGlobalEvents)

	// Kawasan lock/sync poll endpoint
	rg.Get("/inspeksi/kawasan/:kawasanId/poll", authMW, h.PollKawasanEvents)

	// User notifications poll endpoint
	rg.Get("/inspeksi/notifications/poll", authMW, h.PollUserNotifications)
}

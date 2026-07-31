package v1

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/lockusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	pkgkafka "github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterDistributedInspectionRoutes(app *fiber.App, rg fiber.Router, db *gorm.DB, redisClient *redis.Client, producer pkgkafka.EventProducer, jwtManager *jwt.Manager, log *logger.Logger) {
	// Initialize LockManager
	var rClient *redis.Client
	if redisClient != nil {
		rClient = redisClient
	}
	lockMgr := lockusecase.NewLockManager(rClient, log.Logger)

	inspeksiH := handler.NewInspeksiHandler(lockMgr, db, producer, log)
	authMW := middleware.AuthMiddleware(jwtManager)

	// WebSocket Upgrade & Handler (App-level route for WebSocket)
	app.Use("/ws/inspeksi", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/inspeksi/:kawasanId", websocket.New(inspeksiH.WebSocketGrid))

	// SSE Route
	rg.Get("/sse/inspeksi/notifications", inspeksiH.SSENotifications)

	// REST Endpoints under /inspeksi
	insp := rg.Group("/inspeksi", authMW)
	{
		insp.Get("/:kawasanId/status", inspeksiH.GetKawasanStatus)
		insp.Get("/:kawasanId/:aspekId/state", inspeksiH.GetDraftState)

		insp.Post("/:kawasanId/:aspekId/lock", inspeksiH.AcquireLock)
		insp.Put("/:kawasanId/:aspekId/lock/heartbeat", inspeksiH.RenewLock)
		insp.Delete("/:kawasanId/:aspekId/lock", inspeksiH.ReleaseLock)
		insp.Post("/:kawasanId/:aspekId/yield-request", inspeksiH.YieldRequest)

		insp.Put("/:kawasanId/:aspekId", inspeksiH.SaveAspek)
	}
}

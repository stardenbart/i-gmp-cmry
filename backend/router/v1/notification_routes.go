package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/internal/interfaces/http/handlers"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

func RegisterNotificationRoutes(rg fiber.Router, db *gorm.DB, jwtManager *jwt.Manager, log *logger.Logger, cfg *config.Config) {
	// 1 & 2. Repositories + UseCase (in-app bell + Web Push, shared wiring)
	usecase := buildNotificationUseCase(db, cfg)

	// 3. Handlers
	handler := handlers.NewNotificationHandler(usecase)

	// 4. Routes
	notifGroup := rg.Group("/notifications")
	notifGroup.Use(middleware.AuthMiddleware(jwtManager))
	{
		notifGroup.Get("/", handler.GetNotifications)
		notifGroup.Put("/read-all", handler.MarkAllAsRead)
		notifGroup.Put("/:id/read", handler.MarkAsRead)
		notifGroup.Post("/subscribe", handler.Subscribe)
		notifGroup.Post("/unsubscribe", handler.Unsubscribe)
	}
}

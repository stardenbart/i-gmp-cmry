package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/apikey"
	"github.com/monitoring-system/backend/internal/handler/apikeyhandler"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
)

// RegisterAPIKeyRoutes mounts /api/v1/api-keys routes for Admin
func RegisterAPIKeyRoutes(rg fiber.Router, uc apikey.UseCase, jwtManager *jwt.Manager, log *logger.Logger) {
	handler := apikeyhandler.NewAPIKeyHandler(uc, log)

	keysGroup := rg.Group("/api-keys", middleware.AuthMiddleware(jwtManager), middleware.RequireAdmin())

	keysGroup.Post("/", handler.Create)
	keysGroup.Get("/", handler.List)
	keysGroup.Delete("/:id", handler.Revoke)
}

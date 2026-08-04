package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/apikey"
	"github.com/monitoring-system/backend/internal/handler/apikeyhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

// RegisterAPIKeyRoutes mounts /api/v1/api-keys routes for Admin
func RegisterAPIKeyRoutes(rg fiber.Router, db *gorm.DB, uc apikey.UseCase, jwtManager *jwt.Manager, log *logger.Logger) {
	handler := apikeyhandler.NewAPIKeyHandler(uc, log)

	userRepo := authrepo.NewUserRepository(db)
	keysGroup := rg.Group("/api-keys", middleware.AuthMiddleware(jwtManager), middleware.PlantScopeMiddleware(userRepo), middleware.RequireAdmin())

	keysGroup.Post("/", handler.Create)
	keysGroup.Get("/", handler.List)
	keysGroup.Delete("/:id", handler.Revoke)
}

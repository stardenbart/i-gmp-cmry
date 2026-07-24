package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/apikey"
	"github.com/monitoring-system/backend/internal/handler/publichandler"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

// RegisterPowerBIRoutes mounts /api/v1/public/powerbi routes
func RegisterPowerBIRoutes(rg fiber.Router, db *gorm.DB, apikeyUC apikey.UseCase, cryptoSvc *crypto.Service, log *logger.Logger) {
	handler := publichandler.NewPowerBIHandler(db, cryptoSvc, log)

	pbiGroup := rg.Group("/public/powerbi", middleware.APIKeyBearerMiddleware(apikeyUC))

	pbiGroup.Get("/data", handler.GetAllData)
}

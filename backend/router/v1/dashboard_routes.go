package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/dashboardhandler"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

func RegisterDashboardRoutes(rg fiber.Router, db *gorm.DB, cryptoSvc *crypto.Service, jwtManager *jwt.Manager, log *logger.Logger) {
	handler := dashboardhandler.NewDashboardHandler(db, log, cryptoSvc)

	dashboardGroup := rg.Group("/dashboard")
	// For testing, let's allow without JWT first, or require it
	// If you require JWT: dashboardGroup.Use(middleware.JWTMiddleware(jwtManager))
	
	dashboardGroup.Get("/stats", handler.GetStats)
	dashboardGroup.Get("/preview-export", handler.GetPreviewExport)
	dashboardGroup.Get("/export", handler.ExportStats)
	dashboardGroup.Post("/export/template", handler.UploadTemplate)
	dashboardGroup.Get("/auditor-detail", handler.GetAuditorDetail)
	dashboardGroup.Get("/pic-detail", handler.GetPICDetail)
}

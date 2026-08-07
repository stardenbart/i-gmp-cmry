package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/dashboardhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

func RegisterDashboardRoutes(rg fiber.Router, db *gorm.DB, cryptoSvc *crypto.Service, jwtManager *jwt.Manager, log *logger.Logger) {
	userRepo := authrepo.NewUserRepository(db)
	handler := dashboardhandler.NewDashboardHandler(db, log, cryptoSvc)

	authMW := middleware.AuthMiddleware(jwtManager)
	plantScopeMW := middleware.PlantScopeMiddleware(userRepo)

	dashboardGroup := rg.Group("/dashboard", authMW, plantScopeMW)
	
	dashboardGroup.Get("/stats", handler.GetStats, plantScopeMW)
	dashboardGroup.Get("/wowr-report", handler.GetWOWRReport, plantScopeMW)
	dashboardGroup.Get("/preview-export", handler.GetPreviewExport, plantScopeMW)
	dashboardGroup.Get("/export", handler.ExportStats, plantScopeMW)
	dashboardGroup.Post("/export/template", handler.UploadTemplate, plantScopeMW)
	dashboardGroup.Get("/auditor-detail", handler.GetAuditorDetail, plantScopeMW)
	dashboardGroup.Get("/pic-detail", handler.GetPICDetail, plantScopeMW)
}

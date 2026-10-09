package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/dashboardhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/dashboardrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

func RegisterDashboardRoutes(rg fiber.Router, db *gorm.DB, cryptoSvc *crypto.Service, jwtManager *jwt.Manager, log *logger.Logger, appBaseURL string) {
	userRepo := authrepo.NewUserRepository(db)
	layoutRepo := dashboardrepo.NewDashboardLayoutRepository(db)
	settingRepo := masterrepo.NewSettingRepository(db)
	handler := dashboardhandler.NewDashboardHandler(db, log, cryptoSvc, layoutRepo, settingRepo).WithAppBaseURL(appBaseURL)

	authMW := middleware.AuthMiddleware(jwtManager)
	plantScopeMW := middleware.PlantScopeMiddleware(userRepo)

	dashboardGroup := rg.Group("/dashboard", authMW, plantScopeMW)

	dashboardGroup.Get("/stats", handler.GetStats, plantScopeMW)
	dashboardGroup.Get("/trend", handler.GetTrend, plantScopeMW)
	dashboardGroup.Get("/wowr-report", handler.GetWOWRReport, plantScopeMW)
	dashboardGroup.Get("/wowr-report/export", handler.ExportWOWRReport, plantScopeMW)
	dashboardGroup.Get("/preview-export", handler.GetPreviewExport, plantScopeMW)
	dashboardGroup.Get("/export", handler.ExportStats, plantScopeMW)
	dashboardGroup.Post("/export/template", handler.UploadTemplate, plantScopeMW)
	dashboardGroup.Get("/auditor-detail", handler.GetAuditorDetail, plantScopeMW)
	dashboardGroup.Get("/pic-detail", handler.GetPICDetail, plantScopeMW)

	// Per-user dashboard widget layout (Fase 1: checklist + order; no
	// plant scoping needed — this is a personal preference, not plant data).
	dashboardGroup.Get("/layout", handler.GetLayout)
	dashboardGroup.Put("/layout", handler.SaveLayout)
}

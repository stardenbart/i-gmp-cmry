package v1

import (
	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/kpisharehandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/analyticsrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/dashboardrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/kpisharerepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/analyticsusecase"
	"github.com/monitoring-system/backend/internal/usecase/kpishareusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

// RegisterKPIShareRoutes must run before RegisterAuthRoutes because the
// latter mounts a catch-all authenticated group on the v1 router.
func RegisterKPIShareRoutes(rg fiber.Router, db *gorm.DB, jwtManager *jwt.Manager, log *logger.Logger, activity logdomain.ActivityLogUseCase) {
	layoutRepo := dashboardrepo.NewDashboardLayoutRepository(db)
	analyticsRepo := analyticsrepo.NewRepository(db)
	analyticsService := analyticsusecase.NewQueryService(db, analyticsRepo, log)
	shareRepo := kpisharerepo.New(db)
	shareService := kpishareusecase.New(shareRepo, layoutRepo, analyticsService)
	handler := kpisharehandler.New(shareService, activity)

	public := rg.Group("/public/kpi/:token", middleware.APIRateLimiter())
	public.Get("/bootstrap", handler.Bootstrap)
	public.Post("/query", handler.CustomQuery)

	userRepo := authrepo.NewUserRepository(db)
	management := rg.Group("/dashboard/kpi/shares", middleware.AuthMiddleware(jwtManager), middleware.PlantScopeMiddleware(userRepo), middleware.RequireAdmin())
	management.Post("/", handler.Create)
	management.Get("/", handler.List)
	management.Delete("/:shareId", handler.Revoke)
	management.Post("/:shareId/rotate", handler.Rotate)
}

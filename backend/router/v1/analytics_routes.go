package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/analytichandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/analyticsrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/analyticsusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

// RegisterAnalyticsRoutes wires the Custom KPI Visualization Builder's
// semantic-layer endpoints (see backend/internal/domain/analytics and the
// approved plan). Mounted under the same authMW+plantScopeMW pattern as
// RegisterDashboardRoutes, since every analytics query is area-scoped like
// the rest of the dashboard.
func RegisterAnalyticsRoutes(rg fiber.Router, db *gorm.DB, jwtManager *jwt.Manager, log *logger.Logger) {
	userRepo := authrepo.NewUserRepository(db)
	repo := analyticsrepo.NewRepository(db)
	service := analyticsusecase.NewQueryService(db, repo, log)
	handler := analytichandler.NewAnalyticsHandler(service)

	authMW := middleware.AuthMiddleware(jwtManager)
	plantScopeMW := middleware.PlantScopeMiddleware(userRepo)

	analyticsGroup := rg.Group("/analytics", authMW, plantScopeMW)
	analyticsGroup.Get("/catalog", handler.GetCatalog)
	analyticsGroup.Post("/query", handler.RunQuery)
}

package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/logginghandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/loggingrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/loggingusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

// RegisterLoggingRoutes wires logging dependencies and mounts routes.
func RegisterLoggingRoutes(rg fiber.Router, db *gorm.DB, producer kafka.EventProducer, jwtManager *jwt.Manager, log *logger.Logger) {
	loginLogRepo := loggingrepo.NewLoginLogRepository(db)
	actLogRepo := loggingrepo.NewActivityLogRepository(db)
	userRepo := authrepo.NewUserRepository(db)

	loginLogUC := loggingusecase.NewLoginLogUseCase(loginLogRepo)
	actLogUC := loggingusecase.NewActivityLogUseCase(actLogRepo, producer)

	loginLogH := logginghandler.NewLoginLogHandler(loginLogUC)
	actLogH := logginghandler.NewActivityLogHandler(actLogUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	plantScopeMW := middleware.PlantScopeMiddleware(userRepo)
	logs := rg.Group("/logs", authMW, plantScopeMW)
	{
		logs.Get("/login", loginLogH.GetAll)
		logs.Get("/login/:id", loginLogH.GetByID)
		logs.Get("/activity", actLogH.GetAll)
		logs.Get("/activity/:id", actLogH.GetByID)
	}
}

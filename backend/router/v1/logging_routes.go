package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/internal/handler/logginghandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/loggingrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/loggingusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

// RegisterLoggingRoutes wires logging dependencies and mounts routes.
func RegisterLoggingRoutes(rg *gin.RouterGroup, db *gorm.DB, producer kafka.EventProducer, jwtManager *jwt.Manager, log *logger.Logger) {
	loginLogRepo := loggingrepo.NewLoginLogRepository(db)
	actLogRepo := loggingrepo.NewActivityLogRepository(db)

	loginLogUC := loggingusecase.NewLoginLogUseCase(loginLogRepo)
	actLogUC := loggingusecase.NewActivityLogUseCase(actLogRepo, producer)

	loginLogH := logginghandler.NewLoginLogHandler(loginLogUC)
	actLogH := logginghandler.NewActivityLogHandler(actLogUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	logs := rg.Group("/logs", authMW)
	{
		logs.GET("/login", loginLogH.GetAll)
		logs.GET("/login/:id", loginLogH.GetByID)
		logs.GET("/activity", actLogH.GetAll)
		logs.GET("/activity/:id", actLogH.GetByID)
	}
}

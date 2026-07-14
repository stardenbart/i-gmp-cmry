package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/internal/handler/inspectionhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/inspectionrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/inspectionusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

func RegisterInspectionRoutes(rg *gin.RouterGroup, db *gorm.DB, producer kafka.EventProducer, jwtManager *jwt.Manager, log *logger.Logger) {
	headerRepo := inspectionrepo.NewInspectionHeaderRepository(db)
	resultRepo := inspectionrepo.NewInspectionResultRepository(db)

	headerUC := inspectionusecase.NewInspectionHeaderUseCase(headerRepo, producer)
	resultUC := inspectionusecase.NewInspectionResultUseCase(resultRepo)

	headerH := inspectionhandler.NewInspectionHeaderHandler(headerUC)
	resultH := inspectionhandler.NewInspectionResultHandler(resultUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	insp := rg.Group("/inspections", authMW)
	{
		// Inspection Header CRUD
		insp.GET("", headerH.GetAll)
		insp.POST("", headerH.Create)
		insp.GET("/:id", headerH.GetByID)
		insp.PUT("/:id/status", headerH.UpdateStatus)
		insp.DELETE("/:id", headerH.Delete)

		// Inspection Results (nested under header)
		insp.GET("/:id/results", resultH.GetByInspectionID)
		insp.POST("/:id/results/bulk", resultH.BulkSave)
		insp.PUT("/results/:result_id", resultH.Update)
		insp.DELETE("/results/:result_id", resultH.Delete)
	}
}

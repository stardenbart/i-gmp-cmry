package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/inspectionhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/inspectionrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/inspectionusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

func RegisterInspectionRoutes(rg fiber.Router, db *gorm.DB, producer kafka.EventProducer, jwtManager *jwt.Manager, log *logger.Logger) {
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
		insp.Get("", headerH.GetAll)
		insp.Post("", headerH.Create)
		insp.Get("/:id", headerH.GetByID)
		insp.Put("/:id/status", headerH.UpdateStatus)
		insp.Delete("/:id", headerH.Delete)

		// Inspection Results (nested under header)
		insp.Get("/:id/results", resultH.GetByInspectionID)
		insp.Post("/:id/results/bulk", resultH.BulkSave)
		insp.Put("/results/:result_id", resultH.Update)
		insp.Delete("/results/:result_id", resultH.Delete)
	}
}

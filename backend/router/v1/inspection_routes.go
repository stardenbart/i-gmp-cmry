package v1

import (
	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/inspectionhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/inspectionrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/picrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/inspectionusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"gorm.io/gorm"
)

func RegisterInspectionRoutes(rg fiber.Router, db *gorm.DB, producer kafka.EventProducer, mailer mail.Mailer, jwtManager *jwt.Manager, log *logger.Logger, actLogUC logdomain.ActivityLogUseCase) {
	headerRepo := inspectionrepo.NewInspectionHeaderRepository(db)
	resultRepo := inspectionrepo.NewInspectionResultRepository(db)

	detailKawasanRepo := masterrepo.NewDetailKawasanRepository(db)
	kawasanRepo := masterrepo.NewKawasanRepository(db)
	picRepo := picrepo.NewPICMappingRepository(db)
	authRepo := authrepo.NewUserRepository(db)
	settingRepo := masterrepo.NewSettingRepository(db)

	emailNotifier := inspectionusecase.NewInspectionEmailNotifier(mailer, picRepo, authRepo, headerRepo, settingRepo)

	headerUC := inspectionusecase.NewInspectionHeaderUseCase(headerRepo, producer, detailKawasanRepo, kawasanRepo, emailNotifier)
	resultUC := inspectionusecase.NewInspectionResultUseCase(resultRepo)

	// Filter
	filterRepo := inspectionrepo.NewInspectionFilterRepository(db)
	filterUC := inspectionusecase.NewInspectionFilterUseCase(filterRepo)

	headerH := inspectionhandler.NewInspectionHeaderHandler(headerUC, resultUC)
	resultH := inspectionhandler.NewInspectionResultHandler(resultUC)
	filterH := inspectionhandler.NewInspectionFilterHandler(filterUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)
	plantScopeMW := middleware.PlantScopeMiddleware(authRepo)

	// Analytics Route
	rg.Get("/analytics/inspections-trend", authMW, actLogMW, plantScopeMW, headerH.GetTrend)

	insp := rg.Group("/inspections", authMW, actLogMW, plantScopeMW)
	{
		// Inspection Header CRUD
		insp.Get("", headerH.GetAll)
		insp.Post("", headerH.Create)
		// Filter (must be before /:id)
		insp.Get("/filter", filterH.GetFiltered)
		// Area Status
		insp.Get("/area/:areaId/status", headerH.GetAreaStatus)

		insp.Get("/:id/checklist", headerH.GetChecklist)
		insp.Get("/:id", headerH.GetByID)
		insp.Get("/:id/export", headerH.ExportExcel)
		insp.Put("/:id/status", headerH.UpdateStatus)
		insp.Delete("/:id", headerH.Delete)

		// Inspection Results (nested under header)
		insp.Get("/:id/results", resultH.GetByInspectionID)
		insp.Post("/:id/results/bulk", resultH.BulkSave)
		insp.Put("/results/:result_id", resultH.Update)
		insp.Delete("/results/:result_id", resultH.Delete)
	}
}

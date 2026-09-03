package v1

import (
	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/uploadhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/inspectionrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/uploadrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/inspectionusecase"
	"github.com/monitoring-system/backend/internal/usecase/uploadusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/storage"
	"gorm.io/gorm"
)

// RegisterUploadRoutes registers the upload routes
func RegisterUploadRoutes(
	rg fiber.Router,
	db *gorm.DB,
	minioStorage *storage.MinioStorage,
	producer kafka.EventProducer,
	jwtManager *jwt.Manager,
	log *logger.Logger,
	actLogUC logdomain.ActivityLogUseCase,
) {
	uploadRepo := uploadrepo.NewUploadRepository(db)

	docxProc := uploadusecase.NewDOCXProcessor()
	imageProc := uploadusecase.NewImageProcessor(producer)

	uploadUC := uploadusecase.NewUploadUseCase(minioStorage, docxProc, imageProc, uploadRepo)

	// Wired only to authorize an upload's inspection_id against the
	// caller's plant scope (GetByIDScoped) — the fuller instance used by
	// RegisterInspectionRoutes carries extra dependencies (email
	// notifications, period-cutoff resolution) that read/scope-check paths
	// never touch, so nil is fine here.
	headerUC := inspectionusecase.NewInspectionHeaderUseCase(
		inspectionrepo.NewInspectionHeaderRepository(db),
		nil,
		nil,
		nil,
		masterrepo.NewAreaRepository(db),
		nil,
	)

	uploadH := uploadhandler.NewUploadHandler(uploadUC, headerUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)
	plantScopeMW := middleware.PlantScopeMiddleware(authrepo.NewUserRepository(db))

	uploads := rg.Group("/uploads", authMW, actLogMW, plantScopeMW)
	{
		uploads.Post("/file", uploadH.Upload)
		uploads.Get("/:id/status", uploadH.GetStatus)
		uploads.Get("/inspection/:inspectionId", uploadH.GetByInspection)
	}
}

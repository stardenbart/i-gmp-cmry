package v1

import (
	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/uploadhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/uploadrepo"
	"github.com/monitoring-system/backend/internal/middleware"
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

	uploadH := uploadhandler.NewUploadHandler(uploadUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)

	uploads := rg.Group("/uploads", authMW, actLogMW)
	{
		uploads.Post("/file", uploadH.Upload)
		uploads.Get("/:id/status", uploadH.GetStatus)
		uploads.Get("/inspection/:inspectionId", uploadH.GetByInspection)
	}
}

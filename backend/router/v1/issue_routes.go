package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/internal/handler/issuehandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/issuerepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/issueusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/storage"
	"gorm.io/gorm"
)

func RegisterIssueRoutes(rg *gin.RouterGroup, db *gorm.DB, minioStorage *storage.MinioStorage, producer kafka.EventProducer, mailer mail.Mailer, jwtManager *jwt.Manager, log *logger.Logger) {
	issueRepo := issuerepo.NewIssueRepository(db)
	photoRepo := issuerepo.NewIssuePhotoRepository(db)
	userRepo := authrepo.NewUserRepository(db)
	settingRepo := masterrepo.NewSettingRepository(db)

	issueUC := issueusecase.NewIssueUseCase(issueRepo, producer, mailer, userRepo, settingRepo)
	photoUC := issueusecase.NewIssuePhotoUseCase(photoRepo, minioStorage)

	issueH := issuehandler.NewIssueHandler(issueUC)
	photoH := issuehandler.NewIssuePhotoHandler(photoUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	issues := rg.Group("/issues", authMW)
	{
		// Issue CRUD
		issues.GET("", issueH.GetAll)
		issues.POST("", issueH.Create)
		issues.GET("/:id", issueH.GetByID)
		issues.PUT("/:id", issueH.Update)
		issues.DELETE("/:id", issueH.Delete)

		// Issue Photos (initial + follow-up)
		issues.GET("/:id/photos", photoH.GetByIssueID)
		issues.POST("/:id/photos/upload", photoH.Upload)
		issues.DELETE("/photos/:photo_id", photoH.Delete)
	}
}

package v1

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/issuehandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/issuerepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/issueusecase"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/storage"
	"gorm.io/gorm"
)

func RegisterIssueRoutes(rg fiber.Router, db *gorm.DB, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, producer kafka.EventProducer, mailer mail.Mailer, jwtManager *jwt.Manager, log *logger.Logger, actLogUC logdomain.ActivityLogUseCase) {
	issueRepo := issuerepo.NewIssueRepository(db)
	photoRepo := issuerepo.NewIssuePhotoRepository(db)
	userRepo := authrepo.NewUserRepository(db)
	settingRepo := masterrepo.NewSettingRepository(db)

	issueDelegateRepo := issuerepo.NewIssueDelegateRepository(db)
	issueUC := issueusecase.NewIssueUseCase(issueRepo, producer, mailer, userRepo, settingRepo, issueDelegateRepo, cryptoSvc)
	issueDelegateUC := issueusecase.NewIssueDelegateUseCase(issueDelegateRepo)
	photoUC := issueusecase.NewIssuePhotoUseCase(photoRepo, minioStorage, cryptoSvc)

	// Start Background Worker (interval 1 hour)
	issueusecase.StartAutoApproveWorker(context.Background(), issueUC, 1*time.Hour)

	issueH := issuehandler.NewIssueHandler(issueUC)
	issueDelegateH := issuehandler.NewIssueDelegateHandler(issueDelegateUC)
	photoH := issuehandler.NewIssuePhotoHandler(photoUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)
	issues := rg.Group("/issues", authMW, actLogMW)
	{
		// Issue CRUD
		issues.Get("", issueH.GetAll)
		issues.Post("", issueH.Create)
		issues.Get("/:id", issueH.GetByID)
		issues.Put("/:id", issueH.Update)
		issues.Put("/:id/extend-deadline", issueH.ExtendDueDate)
		issues.Delete("/:id", issueH.Delete)

		// Issue Delegates
		issues.Post("/:id/delegates", issueDelegateH.AddDelegate)
		issues.Delete("/:id/delegates/:user_id", issueDelegateH.RemoveDelegate)
		issues.Get("/:id/delegates", issueDelegateH.GetDelegates)

		// Issue Photos (initial + follow-up)
		issues.Get("/:id/photos", photoH.GetByIssueID)
		issues.Post("/:id/photos/upload", photoH.Upload)
		issues.Delete("/photos/:photo_id", photoH.Delete)
	}
}

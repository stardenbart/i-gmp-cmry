package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/config"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/inspectionhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/inspectionrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/issuerepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/picrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/uploadrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/authusecase"
	"github.com/monitoring-system/backend/internal/usecase/inspectionusecase"
	"github.com/monitoring-system/backend/internal/usecase/issueusecase"
	"github.com/monitoring-system/backend/internal/usecase/uploadusecase"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/storage"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterInspectionRoutes(rg fiber.Router, db *gorm.DB, producer kafka.EventProducer, mailer mail.Mailer, jwtManager *jwt.Manager, log *logger.Logger, actLogUC logdomain.ActivityLogUseCase, minioStorage *storage.MinioStorage, rdb *redis.Client, cryptoSvc *crypto.Service, cfg *config.Config) {
	headerRepo := inspectionrepo.NewInspectionHeaderRepository(db)
	resultRepo := inspectionrepo.NewInspectionResultRepository(db)

	detailKawasanRepo := masterrepo.NewDetailKawasanRepository(db)
	kawasanRepo := masterrepo.NewKawasanRepository(db)
	picRepo := picrepo.NewPICMappingRepository(db)
	authRepo := authrepo.NewUserRepository(db)
	settingRepo := masterrepo.NewSettingRepository(db)
	notificationUC := buildNotificationUseCase(db, cfg)

	emailNotifier := inspectionusecase.NewInspectionEmailNotifier(mailer, picRepo, authRepo, headerRepo, settingRepo, notificationUC)

	headerUC := inspectionusecase.NewInspectionHeaderUseCase(headerRepo, producer, detailKawasanRepo, kawasanRepo, emailNotifier)

	// Build IssueUseCase for auto-sync Issues on BulkSave
	issueRepo := issuerepo.NewIssueRepository(db)
	issuePhotoRepo := issuerepo.NewIssuePhotoRepository(db)
	issueHEIRepo := issuerepo.NewIssueHEIRepository(db)
	issueDelegateRepo := issuerepo.NewIssueDelegateRepository(db)
	issueUC := issueusecase.NewIssueUseCase(issueRepo, issuePhotoRepo, issueHEIRepo, minioStorage, producer, mailer, authRepo, settingRepo, issueDelegateRepo, cryptoSvc, rdb, notificationUC)

	resultUC := inspectionusecase.NewInspectionResultUseCase(resultRepo, issueUC, headerUC)

	// Filter
	filterRepo := inspectionrepo.NewInspectionFilterRepository(db)
	filterUC := inspectionusecase.NewInspectionFilterUseCase(filterRepo)

	// Build uploadUC untuk FinalizeInspectionPhotos saat inspeksi Completed
	fileUploadRepo := uploadrepo.NewUploadRepository(db)
	imageProc := uploadusecase.NewImageProcessor(nil) // Kafka producer opsional
	docxProc := uploadusecase.NewDOCXProcessor()
	uploadUC := uploadusecase.NewUploadUseCase(minioStorage, docxProc, imageProc, fileUploadRepo)

	headerH := inspectionhandler.NewInspectionHeaderHandler(headerUC, resultUC, uploadUC, rdb)
	resultH := inspectionhandler.NewInspectionResultHandler(resultUC)
	filterH := inspectionhandler.NewInspectionFilterHandler(filterUC)

	rpRepo := authrepo.NewRolePermissionRepository(db)
	rpUC := authusecase.NewRolePermissionUseCase(rpRepo)
	upRepo := authrepo.NewUserPermissionRepository(db)
	upUC := authusecase.NewUserPermissionUseCase(upRepo)

	permRead := middleware.PermissionMiddleware(rpUC, upUC, "MOD-INSP", "READ")
	permCreate := middleware.PermissionMiddleware(rpUC, upUC, "MOD-INSP", "CREATE")
	permUpdate := middleware.PermissionMiddleware(rpUC, upUC, "MOD-INSP", "UPDATE")
	permApprove := middleware.PermissionMiddleware(rpUC, upUC, "MOD-INSP", "APPROVE")
	permExport := middleware.PermissionMiddleware(rpUC, upUC, "MOD-INSP", "EXPORT")

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)
	plantScopeMW := middleware.PlantScopeMiddleware(authRepo)

	// Analytics Route
	rg.Get("/analytics/inspections-trend", authMW, actLogMW, plantScopeMW, permRead, headerH.GetTrend)

	insp := rg.Group("/inspections", authMW, actLogMW, plantScopeMW)
	{
		// Inspection Header CRUD
		insp.Get("", permRead, headerH.GetAll)
		insp.Post("", permCreate, headerH.Create)
		// Filter (must be before /:id)
		insp.Get("/filter", permRead, filterH.GetFiltered)
		// Area Status
		insp.Get("/area/:areaId/status", permRead, headerH.GetAreaStatus)

		insp.Get("/:id/checklist", permRead, headerH.GetChecklist)
		insp.Get("/:id", permRead, headerH.GetByID)
		insp.Get("/:id/export", permExport, headerH.ExportExcel)
		insp.Put("/:id/status", permApprove, headerH.UpdateStatus)
		insp.Delete("/:id", permUpdate, headerH.Delete)

		// Inspection Results (nested under header)
		insp.Get("/:id/results", permRead, resultH.GetByInspectionID)
		insp.Post("/:id/results/bulk", permUpdate, resultH.BulkSave)
		insp.Put("/results/:result_id", permUpdate, resultH.Update)
		insp.Delete("/results/:result_id", permUpdate, resultH.Delete)
	}
}

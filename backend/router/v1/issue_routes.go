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
	"github.com/monitoring-system/backend/internal/usecase/authusecase"
	"github.com/monitoring-system/backend/internal/usecase/issueusecase"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/storage"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterIssueRoutes(rg fiber.Router, db *gorm.DB, redisClient *redis.Client, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, producer kafka.EventProducer, mailer mail.Mailer, jwtManager *jwt.Manager, log *logger.Logger, actLogUC logdomain.ActivityLogUseCase) {
	issueRepo := issuerepo.NewIssueRepository(db)
	photoRepo := issuerepo.NewIssuePhotoRepository(db)
	heiRepo := issuerepo.NewIssueHEIRepository(db)
	userRepo := authrepo.NewUserRepository(db)
	settingRepo := masterrepo.NewSettingRepository(db)
	issueDelegateRepo := issuerepo.NewIssueDelegateRepository(db)

	issueUC := issueusecase.NewIssueUseCase(issueRepo, photoRepo, heiRepo, minioStorage, producer, mailer, userRepo, settingRepo, issueDelegateRepo, cryptoSvc, redisClient)
	issueDelegateUC := issueusecase.NewIssueDelegateUseCase(issueDelegateRepo)
	photoUC := issueusecase.NewIssuePhotoUseCase(photoRepo, issueRepo, minioStorage, cryptoSvc, redisClient, producer)

	// Filter usecases
	issueFilterRepo := issuerepo.NewIssueFilterRepository(db)
	issueFilterUC := issueusecase.NewIssueFilterUseCase(issueFilterRepo, cryptoSvc)
	followupFilterRepo := issuerepo.NewFollowupFilterRepository(db)
	followupFilterUC := issueusecase.NewFollowupFilterUseCase(followupFilterRepo, cryptoSvc)

	// Start Background Worker (interval 1 hour)
	issueusecase.StartAutoApproveWorker(context.Background(), issueUC, 1*time.Hour)

	issueH := issuehandler.NewIssueHandler(issueUC)
	issueDelegateH := issuehandler.NewIssueDelegateHandler(issueDelegateUC)
	photoH := issuehandler.NewIssuePhotoHandler(photoUC)
	issueFilterH := issuehandler.NewIssueFilterHandler(issueFilterUC)
	followupFilterH := issuehandler.NewFollowupFilterHandler(followupFilterUC)

	rpRepo := authrepo.NewRolePermissionRepository(db)
	rpUC := authusecase.NewRolePermissionUseCase(rpRepo)
	upRepo := authrepo.NewUserPermissionRepository(db)
	upUC := authusecase.NewUserPermissionUseCase(upRepo)
	permReadIss := middleware.PermissionMiddleware(rpUC, upUC, "MOD-ISS", "READ")
	permCreateIss := middleware.PermissionMiddleware(rpUC, upUC, "MOD-ISS", "CREATE")
	permUpdateIss := middleware.PermissionMiddleware(rpUC, upUC, "MOD-ISS", "UPDATE")

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)
	plantScopeMW := middleware.PlantScopeMiddleware(userRepo)
	issues := rg.Group("/issues", authMW, actLogMW, plantScopeMW)
	{
		// Issue CRUD
		issues.Get("", permReadIss, issueH.GetAll)
		issues.Post("", permCreateIss, issueH.Create)
		// Static filter routes MUST come before /:id
		issues.Get("/filter", permReadIss, issueFilterH.GetFiltered)
		issues.Get("/followup/filter", permReadIss, followupFilterH.GetFiltered)
		issues.Post("/close-by-result", permUpdateIss, issueH.CloseByResult)
		issues.Get("/by-result/:result_id", permReadIss, issueH.GetByResultID)

		issues.Get("/:id", permReadIss, issueH.GetByID)
		issues.Put("/:id", permUpdateIss, issueH.Update)
		issues.Put("/:id/extend-deadline", permUpdateIss, issueH.ExtendDueDate)
		issues.Delete("/:id", permUpdateIss, issueH.Delete)

		// Issue Delegates
		issues.Post("/:id/delegates", permUpdateIss, issueDelegateH.AddDelegate)
		issues.Delete("/:id/delegates/:user_id", permUpdateIss, issueDelegateH.RemoveDelegate)
		issues.Get("/:id/delegates", permReadIss, issueDelegateH.GetDelegates)

		// Issue Photos (initial + follow-up)
		issues.Get("/:id/photos", permReadIss, photoH.GetByIssueID)
		issues.Post("/:id/photos/upload", permUpdateIss, photoH.Upload)
		issues.Put("/photos/:photo_id", permUpdateIss, photoH.Update)
		issues.Put("/photos/:photo_id/hei", permUpdateIss, photoH.UpdateHEI)
		issues.Put("/photos/:photo_id/wowr", permUpdateIss, photoH.UpdateWOWR)
		issues.Delete("/photos/:photo_id", permUpdateIss, photoH.Delete)
	}

	// ── HEI (Habit, Equipment, Infrastructure) Master Routes ─────────────────
	habitRepo := issuerepo.NewHabitRepository(db)
	equipmentRepo := issuerepo.NewEquipmentRepository(db)
	infraRepo := issuerepo.NewInfrastructureRepository(db)

	habitUC := issueusecase.NewHabitUseCase(habitRepo)
	equipmentUC := issueusecase.NewEquipmentUseCase(equipmentRepo)
	infraUC := issueusecase.NewInfrastructureUseCase(infraRepo)

	heiH := issuehandler.NewHEIHandler(habitUC, equipmentUC, infraUC)

	permReadMstr := middleware.PermissionMiddleware(rpUC, upUC, "MOD-MSTR", "READ")
	permCreateMstr := middleware.PermissionMiddleware(rpUC, upUC, "MOD-MSTR", "CREATE")
	permUpdateMstr := middleware.PermissionMiddleware(rpUC, upUC, "MOD-MSTR", "UPDATE")
	permDeleteMstr := middleware.PermissionMiddleware(rpUC, upUC, "MOD-MSTR", "DELETE")

	masterGroup := rg.Group("/master", authMW, actLogMW)
	{
		habits := masterGroup.Group("/habits")
		habits.Get("", permReadMstr, heiH.GetAllHabits)
		habits.Post("", permCreateMstr, heiH.CreateHabit)
		habits.Get("/:id", permReadMstr, heiH.GetHabitByID)
		habits.Put("/:id", permUpdateMstr, heiH.UpdateHabit)
		habits.Delete("/:id", permDeleteMstr, heiH.DeleteHabit)

		equipments := masterGroup.Group("/equipments")
		equipments.Get("", permReadMstr, heiH.GetAllEquipments)
		equipments.Post("", permCreateMstr, heiH.CreateEquipment)
		equipments.Get("/:id", permReadMstr, heiH.GetEquipmentByID)
		equipments.Put("/:id", permUpdateMstr, heiH.UpdateEquipment)
		equipments.Delete("/:id", permDeleteMstr, heiH.DeleteEquipment)

		infras := masterGroup.Group("/infrastructures")
		infras.Get("", permReadMstr, heiH.GetAllInfrastructures)
		infras.Post("", permCreateMstr, heiH.CreateInfrastructure)
		infras.Get("/:id", permReadMstr, heiH.GetInfrastructureByID)
		infras.Put("/:id", permUpdateMstr, heiH.UpdateInfrastructure)
		infras.Delete("/:id", permDeleteMstr, heiH.DeleteInfrastructure)
	}
}

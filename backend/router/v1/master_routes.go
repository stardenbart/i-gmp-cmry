package v1

import (
	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/auth"
	"github.com/monitoring-system/backend/internal/handler/issuehandler"
	"github.com/monitoring-system/backend/internal/handler/masterhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/issuerepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/authusecase"
	"github.com/monitoring-system/backend/internal/usecase/issueusecase"
	"github.com/monitoring-system/backend/internal/usecase/masterusecase"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/storage"
	"gorm.io/gorm"
)

// RegisterMasterRoutes wires master data dependencies and mounts routes.
func RegisterMasterRoutes(rg fiber.Router, db *gorm.DB, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, jwtManager *jwt.Manager, log *logger.Logger, actLogUC logdomain.ActivityLogUseCase) {
	// ── Wire dependencies ──────────────────────────────────────────────
	deptRepo := masterrepo.NewDepartmentRepository(db)
	plantRepo := masterrepo.NewPlantRepository(db)
	areaRepo := masterrepo.NewAreaRepository(db)
	kawasanRepo := masterrepo.NewKawasanRepository(db)
	dkRepo := masterrepo.NewDetailKawasanRepository(db)
	aspekRepo := masterrepo.NewAspekRepository(db)
	detailRepo := masterrepo.NewDetailRepository(db)
	uraianRepo := masterrepo.NewUraianRepository(db)
	roleRepo := authrepo.NewRoleRepository(db)
	moduleRepo := authrepo.NewModuleRepository(db)

	habitRepo := issuerepo.NewHabitRepository(db)
	equipRepo := issuerepo.NewEquipmentRepository(db)
	infraRepo := issuerepo.NewInfrastructureRepository(db)

	deptUC := masterusecase.NewDepartmentUseCase(deptRepo)
	plantUC := masterusecase.NewPlantUseCase(plantRepo)
	areaUC := masterusecase.NewAreaUseCase(areaRepo)
	kawasanUC := masterusecase.NewKawasanUseCase(kawasanRepo)
	dkUC := masterusecase.NewDetailKawasanUseCase(dkRepo)
	aspekUC := masterusecase.NewAspekUseCase(aspekRepo)
	detailUC := masterusecase.NewDetailUseCase(detailRepo)
	uraianUC := masterusecase.NewUraianUseCase(uraianRepo)
	roleUC := authusecase.NewRoleUseCase(roleRepo)
	moduleUC := authusecase.NewModuleUseCase(moduleRepo)

	habitUC := issueusecase.NewHabitUseCase(habitRepo)
	equipUC := issueusecase.NewEquipmentUseCase(equipRepo)
	infraUC := issueusecase.NewInfrastructureUseCase(infraRepo)

	rolePermRepo := authrepo.NewRolePermissionRepository(db)
	rolePermUC := authusecase.NewRolePermissionUseCase(rolePermRepo)
	userPermRepo := authrepo.NewUserPermissionRepository(db)
	userPermUC := authusecase.NewUserPermissionUseCase(userPermRepo)

	settingRepo := masterrepo.NewSettingRepository(db)
	settingUC := masterusecase.NewSettingUseCase(settingRepo, cryptoSvc, minioStorage)
	settingH := masterhandler.NewSettingHandler(settingUC)

	deptH := masterhandler.NewDepartmentHandler(deptUC)
	plantH := masterhandler.NewPlantHandler(plantUC)
	areaH := masterhandler.NewAreaHandler(areaUC)
	kawasanH := masterhandler.NewKawasanHandler(kawasanUC)
	dkH := masterhandler.NewDetailKawasanHandler(dkUC)
	aspekH := masterhandler.NewAspekHandler(aspekUC)
	detailH := masterhandler.NewDetailHandler(detailUC)
	uraianH := masterhandler.NewUraianHandler(uraianUC)
	roleH := auth.NewRoleHandler(roleUC)
	rolePermH := auth.NewRolePermissionHandler(rolePermUC)
	moduleH := auth.NewModuleHandler(moduleUC)
	heiH := issuehandler.NewHEIHandler(habitUC, equipUC, infraUC)

	userRepo := authrepo.NewUserRepository(db)
	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)
	plantScopeMW := middleware.PlantScopeMiddleware(userRepo)
	master := rg.Group("/master", authMW, actLogMW, plantScopeMW)
	{
		permReadMstr := middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-MSTR", "READ")
		permCreateMstr := middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-MSTR", "CREATE")
		permUpdateMstr := middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-MSTR", "UPDATE")
		permDeleteMstr := middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-MSTR", "DELETE")

		requireSuperAdmin := middleware.RequireSuperAdminMiddleware()

		// Plants (CRUD modification exclusively accessible by SuperAdmin, GET list readable for plant filter dropdowns)
		plant := master.Group("/plants")
		plant.Get("", plantH.GetAll)
		plant.Post("", requireSuperAdmin, plantH.Create)
		plant.Get("/:id", plantH.GetByID)
		plant.Put("/:id", requireSuperAdmin, plantH.Update)
		plant.Delete("/:id", requireSuperAdmin, plantH.Delete)

		// Department
		dept := master.Group("/departments")
		dept.Get("", permReadMstr, deptH.GetAll)
		dept.Post("", permCreateMstr, deptH.Create)
		dept.Get("/:id", permReadMstr, deptH.GetByID)
		dept.Put("/:id", permUpdateMstr, deptH.Update)
		dept.Delete("/:id", permDeleteMstr, deptH.Delete)

		// Roles
		permReadRole := middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-ROLE", "READ")
		permUpdateRole := middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-ROLE", "UPDATE")
		roles := master.Group("/roles")
		roles.Get("", permReadRole, roleH.GetAll)
		roles.Post("", permUpdateRole, roleH.Create)
		roles.Get("/:id", permReadRole, roleH.GetByID)
		roles.Put("/:id", permUpdateRole, roleH.Update)
		roles.Delete("/:id", permUpdateRole, roleH.Delete)

		roles.Get("/:id/permissions", rolePermH.GetByRoleID)
		roles.Put("/:id/permissions", permUpdateRole, rolePermH.SetPermissions)

		// Modules (Features & Permissions list)
		modules := master.Group("/modules")
		modules.Get("", permReadRole, moduleH.GetAll)

		// Area
		area := master.Group("/area")
		area.Get("", permReadMstr, areaH.GetAll)
		area.Post("", permCreateMstr, areaH.Create)
		area.Get("/:id", permReadMstr, areaH.GetByID)
		area.Put("/:id", permUpdateMstr, areaH.Update)
		area.Delete("/:id", permDeleteMstr, areaH.Delete)

		// Kawasan (filtered by area)
		kawasan := master.Group("/kawasan")
		kawasan.Get("", permReadMstr, kawasanH.GetAll)
		kawasan.Post("", permCreateMstr, kawasanH.Create)
		kawasan.Get("/:id", permReadMstr, kawasanH.GetByID)
		kawasan.Put("/:id", permUpdateMstr, kawasanH.Update)
		kawasan.Delete("/:id", permDeleteMstr, kawasanH.Delete)

		// Detail Kawasan
		dk := master.Group("/detail-kawasan")
		dk.Get("", permReadMstr, dkH.GetAll)
		dk.Post("", permCreateMstr, dkH.Create)
		dk.Get("/:id", permReadMstr, dkH.GetByID)
		dk.Put("/:id", permUpdateMstr, dkH.Update)
		dk.Delete("/:id", permDeleteMstr, dkH.Delete)

		// Aspek
		aspek := master.Group("/aspek")
		aspek.Get("", permReadMstr, aspekH.GetAll)
		aspek.Post("", permCreateMstr, aspekH.Create)
		aspek.Get("/:id", permReadMstr, aspekH.GetByID)
		aspek.Put("/:id", permUpdateMstr, aspekH.Update)
		aspek.Delete("/:id", permDeleteMstr, aspekH.Delete)

		// Detail Audit
		detail := master.Group("/details")
		detail.Get("", permReadMstr, detailH.GetAll)
		detail.Post("", permCreateMstr, detailH.Create)
		detail.Get("/:id", permReadMstr, detailH.GetByID)
		detail.Put("/:id", permUpdateMstr, detailH.Update)
		detail.Delete("/:id", permDeleteMstr, detailH.Delete)

		// Uraian (checklist items)
		uraian := master.Group("/urain")
		uraian.Get("", permReadMstr, uraianH.GetAll)
		uraian.Post("", permCreateMstr, uraianH.Create)
		uraian.Get("/:id", permReadMstr, uraianH.GetByID)
		uraian.Put("/:id", permUpdateMstr, uraianH.Update)
		uraian.Delete("/:id", permDeleteMstr, uraianH.Delete)

		// Habits
		habits := master.Group("/habits")
		habits.Get("", permReadMstr, heiH.GetAllHabits)
		habits.Post("", permCreateMstr, heiH.CreateHabit)
		habits.Get("/:id", permReadMstr, heiH.GetHabitByID)
		habits.Put("/:id", permUpdateMstr, heiH.UpdateHabit)
		habits.Delete("/:id", permDeleteMstr, heiH.DeleteHabit)

		// Equipments
		equipments := master.Group("/equipments")
		equipments.Get("", permReadMstr, heiH.GetAllEquipments)
		equipments.Post("", permCreateMstr, heiH.CreateEquipment)
		equipments.Get("/:id", permReadMstr, heiH.GetEquipmentByID)
		equipments.Put("/:id", permUpdateMstr, heiH.UpdateEquipment)
		equipments.Delete("/:id", permDeleteMstr, heiH.DeleteEquipment)

		// Infrastructures
		infrastructures := master.Group("/infrastructures")
		infrastructures.Get("", permReadMstr, heiH.GetAllInfrastructures)
		infrastructures.Post("", permCreateMstr, heiH.CreateInfrastructure)
		infrastructures.Get("/:id", permReadMstr, heiH.GetInfrastructureByID)
		infrastructures.Put("/:id", permUpdateMstr, heiH.UpdateInfrastructure)
		infrastructures.Delete("/:id", permDeleteMstr, heiH.DeleteInfrastructure)

		// System Settings
		permReadStng := middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-STNG", "READ")
		permUpdateStng := middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-STNG", "UPDATE")
		settings := master.Group("/settings")
		settings.Get("", permReadStng, settingH.GetAll)
		settings.Get("/:key", permReadStng, settingH.GetByKey)
		settings.Put("/:key", permUpdateStng, settingH.Update)
	}
}

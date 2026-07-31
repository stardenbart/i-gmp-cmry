package v1

import (
	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/auth"
	"github.com/monitoring-system/backend/internal/handler/masterhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/authusecase"
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

	rolePermRepo := authrepo.NewRolePermissionRepository(db)
	rolePermUC := authusecase.NewRolePermissionUseCase(rolePermRepo)

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

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)
	master := rg.Group("/master", authMW, actLogMW)
	{
		// Plants
		plant := master.Group("/plants")
		plant.Get("", plantH.GetAll)
		plant.Post("", plantH.Create)
		plant.Get("/:id", plantH.GetByID)
		plant.Put("/:id", plantH.Update)
		plant.Delete("/:id", plantH.Delete)

		// Department
		dept := master.Group("/departments")
		dept.Get("", deptH.GetAll)
		dept.Post("", deptH.Create)
		dept.Get("/:id", deptH.GetByID)
		dept.Put("/:id", deptH.Update)
		dept.Delete("/:id", deptH.Delete)

		// Roles
		roles := master.Group("/roles")
		roles.Get("", roleH.GetAll)
		roles.Post("", roleH.Create)
		roles.Get("/:id", roleH.GetByID)
		roles.Put("/:id", roleH.Update)
		roles.Delete("/:id", roleH.Delete)

		roles.Get("/:id/permissions", rolePermH.GetByRoleID)
		roles.Put("/:id/permissions", rolePermH.SetPermissions)

		// Modules (Features & Permissions list)
		modules := master.Group("/modules")
		modules.Get("", moduleH.GetAll)

		// Area
		area := master.Group("/area")
		area.Get("", areaH.GetAll)
		area.Post("", areaH.Create)
		area.Get("/:id", areaH.GetByID)
		area.Put("/:id", areaH.Update)
		area.Delete("/:id", areaH.Delete)

		// Kawasan (filtered by area)
		kawasan := master.Group("/kawasan")
		kawasan.Get("", kawasanH.GetAll)
		kawasan.Post("", kawasanH.Create)
		kawasan.Get("/:id", kawasanH.GetByID)
		kawasan.Put("/:id", kawasanH.Update)
		kawasan.Delete("/:id", kawasanH.Delete)

		// Detail Kawasan
		dk := master.Group("/detail-kawasan")
		dk.Get("", dkH.GetAll)
		dk.Post("", dkH.Create)
		dk.Get("/:id", dkH.GetByID)
		dk.Put("/:id", dkH.Update)
		dk.Delete("/:id", dkH.Delete)

		// Aspek
		aspek := master.Group("/aspek")
		aspek.Get("", aspekH.GetAll)
		aspek.Post("", aspekH.Create)
		aspek.Get("/:id", aspekH.GetByID)
		aspek.Put("/:id", aspekH.Update)
		aspek.Delete("/:id", aspekH.Delete)

		// Detail Audit
		detail := master.Group("/details")
		detail.Get("", detailH.GetAll)
		detail.Post("", detailH.Create)
		detail.Get("/:id", detailH.GetByID)
		detail.Put("/:id", detailH.Update)
		detail.Delete("/:id", detailH.Delete)

		// Uraian (checklist items)
		uraian := master.Group("/urain")
		uraian.Get("", uraianH.GetAll)
		uraian.Post("", uraianH.Create)
		uraian.Get("/:id", uraianH.GetByID)
		uraian.Put("/:id", uraianH.Update)
		uraian.Delete("/:id", uraianH.Delete)

		// System Settings (Admin only)
		settings := master.Group("/settings")
		settings.Get("", settingH.GetAll)
		settings.Get("/:key", settingH.GetByKey)
		settings.Put("/:key", settingH.Update)
	}
}

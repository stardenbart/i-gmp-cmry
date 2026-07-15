package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/masterhandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/masterusecase"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/storage"
	"gorm.io/gorm"
)

// RegisterMasterRoutes wires master data dependencies and mounts routes.
func RegisterMasterRoutes(rg fiber.Router, db *gorm.DB, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, jwtManager *jwt.Manager, log *logger.Logger) {
	// ── Wire dependencies ──────────────────────────────────────────────
	deptRepo := masterrepo.NewDepartmentRepository(db)
	areaRepo := masterrepo.NewAreaRepository(db)
	kawasanRepo := masterrepo.NewKawasanRepository(db)
	dkRepo := masterrepo.NewDetailKawasanRepository(db)
	aspekRepo := masterrepo.NewAspekRepository(db)
	detailRepo := masterrepo.NewDetailRepository(db)
	uraianRepo := masterrepo.NewUraianRepository(db)

	deptUC := masterusecase.NewDepartmentUseCase(deptRepo)
	areaUC := masterusecase.NewAreaUseCase(areaRepo)
	kawasanUC := masterusecase.NewKawasanUseCase(kawasanRepo)
	dkUC := masterusecase.NewDetailKawasanUseCase(dkRepo)
	aspekUC := masterusecase.NewAspekUseCase(aspekRepo)
	detailUC := masterusecase.NewDetailUseCase(detailRepo)
	uraianUC := masterusecase.NewUraianUseCase(uraianRepo)

	settingRepo := masterrepo.NewSettingRepository(db)
	settingUC := masterusecase.NewSettingUseCase(settingRepo, cryptoSvc, minioStorage)
	settingH := masterhandler.NewSettingHandler(settingUC)

	deptH := masterhandler.NewDepartmentHandler(deptUC)
	areaH := masterhandler.NewAreaHandler(areaUC)
	kawasanH := masterhandler.NewKawasanHandler(kawasanUC)
	dkH := masterhandler.NewDetailKawasanHandler(dkUC)
	aspekH := masterhandler.NewAspekHandler(aspekUC)
	detailH := masterhandler.NewDetailHandler(detailUC)
	uraianH := masterhandler.NewUraianHandler(uraianUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	master := rg.Group("/master", authMW)
	{
		// Department
		dept := master.Group("/departments")
		dept.Get("", deptH.GetAll)
		dept.Post("", deptH.Create)
		dept.Get("/:id", deptH.GetByID)
		dept.Put("/:id", deptH.Update)
		dept.Delete("/:id", deptH.Delete)

		// Area
		area := master.Group("/areas")
		area.Get("", areaH.GetAll)
		area.Post("", areaH.Create)
		area.Get("/:id", areaH.GetByID)
		area.Put("/:id", areaH.Update)
		area.Delete("/:id", areaH.Delete)

		// Kawasan (filtered by area)
		kawasan := master.Group("/kawasans")
		kawasan.Get("", kawasanH.GetAll)
		kawasan.Post("", kawasanH.Create)
		kawasan.Get("/:id", kawasanH.GetByID)
		kawasan.Put("/:id", kawasanH.Update)
		kawasan.Delete("/:id", kawasanH.Delete)

		// Detail Kawasan
		dk := master.Group("/detail-kawasans")
		dk.Get("", dkH.GetAll)
		dk.Post("", dkH.Create)
		dk.Get("/:id", dkH.GetByID)
		dk.Put("/:id", dkH.Update)
		dk.Delete("/:id", dkH.Delete)

		// Aspek
		aspek := master.Group("/aspeks")
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
		uraian := master.Group("/urains")
		uraian.Get("", uraianH.GetAll)
		uraian.Post("", uraianH.Create)
		uraian.Get("/:id", uraianH.GetByID)
		uraian.Put("/:id", uraianH.Update)
		uraian.Delete("/:id", uraianH.Delete)

		// System Settings (Admin only)
		settings := rg.Group("/settings", authMW)
		settings.Get("", settingH.GetAll)
		settings.Get("/:key", settingH.GetByKey)
		settings.Put("/:key", settingH.Update)
	}
}

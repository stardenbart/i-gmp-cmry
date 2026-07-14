package v1

import (
	"github.com/gin-gonic/gin"
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
func RegisterMasterRoutes(rg *gin.RouterGroup, db *gorm.DB, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, jwtManager *jwt.Manager, log *logger.Logger) {
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
		dept.GET("", deptH.GetAll)
		dept.POST("", deptH.Create)
		dept.GET("/:id", deptH.GetByID)
		dept.PUT("/:id", deptH.Update)
		dept.DELETE("/:id", deptH.Delete)

		// Area
		area := master.Group("/areas")
		area.GET("", areaH.GetAll)
		area.POST("", areaH.Create)
		area.GET("/:id", areaH.GetByID)
		area.PUT("/:id", areaH.Update)
		area.DELETE("/:id", areaH.Delete)

		// Kawasan (filtered by area)
		kawasan := master.Group("/kawasans")
		kawasan.GET("", kawasanH.GetAll)
		kawasan.POST("", kawasanH.Create)
		kawasan.GET("/:id", kawasanH.GetByID)
		kawasan.PUT("/:id", kawasanH.Update)
		kawasan.DELETE("/:id", kawasanH.Delete)

		// Detail Kawasan
		dk := master.Group("/detail-kawasans")
		dk.GET("", dkH.GetAll)
		dk.POST("", dkH.Create)
		dk.GET("/:id", dkH.GetByID)
		dk.PUT("/:id", dkH.Update)
		dk.DELETE("/:id", dkH.Delete)

		// Aspek
		aspek := master.Group("/aspeks")
		aspek.GET("", aspekH.GetAll)
		aspek.POST("", aspekH.Create)
		aspek.GET("/:id", aspekH.GetByID)
		aspek.PUT("/:id", aspekH.Update)
		aspek.DELETE("/:id", aspekH.Delete)

		// Detail Audit
		detail := master.Group("/details")
		detail.GET("", detailH.GetAll)
		detail.POST("", detailH.Create)
		detail.GET("/:id", detailH.GetByID)
		detail.PUT("/:id", detailH.Update)
		detail.DELETE("/:id", detailH.Delete)

		// Uraian (checklist items)
		uraian := master.Group("/urains")
		uraian.GET("", uraianH.GetAll)
		uraian.POST("", uraianH.Create)
		uraian.GET("/:id", uraianH.GetByID)
		uraian.PUT("/:id", uraianH.Update)
		uraian.DELETE("/:id", uraianH.Delete)

		// System Settings (Admin only)
		settings := rg.Group("/settings", authMW)
		settings.GET("", settingH.GetAll)
		settings.GET("/:key", settingH.GetByKey)
		settings.PUT("/:key", settingH.Update)
	}
}

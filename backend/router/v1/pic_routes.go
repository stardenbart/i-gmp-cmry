package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/internal/handler/pichandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/picrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/picusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

func RegisterPICRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.Manager, log *logger.Logger) {
	repo := picrepo.NewPICMappingRepository(db)
	uc := picusecase.NewPICMappingUseCase(repo)
	h := pichandler.NewPICMappingHandler(uc)

	authMW := middleware.AuthMiddleware(jwtManager)
	pic := rg.Group("/pic-mappings", authMW)
	{
		pic.GET("", h.GetAll)
		pic.POST("", h.Create)
		pic.GET("/:id", h.GetByID)
		pic.PUT("/:id", h.Update)
		pic.DELETE("/:id", h.Delete)
	}
}

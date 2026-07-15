package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/pichandler"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/picrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/picusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

func RegisterPICRoutes(rg fiber.Router, db *gorm.DB, jwtManager *jwt.Manager, log *logger.Logger) {
	repo := picrepo.NewPICMappingRepository(db)
	uc := picusecase.NewPICMappingUseCase(repo)
	h := pichandler.NewPICMappingHandler(uc)

	authMW := middleware.AuthMiddleware(jwtManager)
	pic := rg.Group("/pic-mappings", authMW)
	{
		pic.Get("", h.GetAll)
		pic.Post("", h.Create)
		pic.Get("/:id", h.GetByID)
		pic.Put("/:id", h.Update)
		pic.Delete("/:id", h.Delete)
	}
}

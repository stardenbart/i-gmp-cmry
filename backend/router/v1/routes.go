package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/apikeyrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/loggingrepo"
	"github.com/monitoring-system/backend/internal/usecase/apikeyusecase"
	"github.com/monitoring-system/backend/internal/usecase/loggingusecase"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/opensearch"
	"github.com/monitoring-system/backend/pkg/storage"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Register mounts all v1 route groups onto the given RouterGroup.
// Each route file is responsible for wiring its own repositories, use cases, and handlers.
func Register(app *fiber.App, rg fiber.Router, db *gorm.DB, redisClient *redis.Client, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, mailer mail.Mailer, producer kafka.EventProducer, osClient *opensearch.Client, jwtManager *jwt.Manager, log *logger.Logger) {
	// Bootstrap shared ActivityLog UseCase — passed to all route groups for activity tracking
	actLogRepo := loggingrepo.NewActivityLogRepository(db)
	actLogUC := loggingusecase.NewActivityLogUseCase(actLogRepo, producer)

	// Bootstrap APIKey UseCase
	apiKeyRepo := apikeyrepo.NewAPIKeyRepository(db)
	apiKeyUC := apikeyusecase.NewAPIKeyUseCase(apiKeyRepo)

	// NOTE: PowerBI and APIKey public routes MUST be registered BEFORE RegisterAuthRoutes.
	// RegisterAuthRoutes creates a protected := rg.Group("/", authMW) group which in Fiber
	// matches ALL paths. Registering public routes first ensures Fiber matches them first.
	RegisterAPIKeyRoutes(rg, db, apiKeyUC, jwtManager, log)
	RegisterPowerBIRoutes(rg, db, apiKeyUC, cryptoSvc, log)
	RegisterPollingRoutes(app, rg, redisClient, jwtManager)

	RegisterAuthRoutes(rg, db, mailer, jwtManager, log, actLogUC)
	RegisterMasterRoutes(rg, db, minioStorage, cryptoSvc, jwtManager, log, actLogUC)
	RegisterPICRoutes(rg, db, jwtManager, log, actLogUC)
	RegisterInspectionRoutes(rg, db, producer, mailer, jwtManager, log, actLogUC)
	RegisterDistributedInspectionRoutes(app, rg, db, redisClient, producer, jwtManager, log)
	RegisterIssueRoutes(rg, db, redisClient, minioStorage, cryptoSvc, producer, mailer, jwtManager, log, actLogUC)
	RegisterLoggingRoutes(rg, db, producer, jwtManager, log)
	RegisterSearchRoutes(rg, osClient, jwtManager, log)
	RegisterUploadRoutes(rg, db, minioStorage, producer, jwtManager, log, actLogUC)
	RegisterNotificationRoutes(rg, db, jwtManager, log)
	RegisterDashboardRoutes(rg, db, cryptoSvc, jwtManager, log)
}

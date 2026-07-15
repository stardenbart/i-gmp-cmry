package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/opensearch"
	"github.com/monitoring-system/backend/pkg/storage"
	"gorm.io/gorm"
)

// Register mounts all v1 route groups onto the given RouterGroup.
// Each route file is responsible for wiring its own repositories, use cases, and handlers.
func Register(rg fiber.Router, db *gorm.DB, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, mailer mail.Mailer, producer kafka.EventProducer, osClient *opensearch.Client, jwtManager *jwt.Manager, log *logger.Logger) {
	RegisterAuthRoutes(rg, db, mailer, jwtManager, log)
	RegisterMasterRoutes(rg, db, minioStorage, cryptoSvc, jwtManager, log)
	RegisterPICRoutes(rg, db, jwtManager, log)
	RegisterInspectionRoutes(rg, db, producer, jwtManager, log)
	RegisterIssueRoutes(rg, db, minioStorage, producer, mailer, jwtManager, log)
	RegisterLoggingRoutes(rg, db, producer, jwtManager, log)
	RegisterSearchRoutes(rg, osClient, jwtManager, log)
}

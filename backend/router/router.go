package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/monitoring-system/backend/config"
	_ "github.com/monitoring-system/backend/docs"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/opensearch"
	"github.com/monitoring-system/backend/pkg/storage"
	"gorm.io/gorm"

	// Handler imports (will be wired during full implementation)
	v1 "github.com/monitoring-system/backend/router/v1"
)

// Setup initialises Fiber, applies global middleware, and registers all API routes.
func Setup(cfg *config.Config, db *gorm.DB, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, mailer mail.Mailer, producer kafka.EventProducer, osClient *opensearch.Client, log *logger.Logger) *fiber.App {
	// Fiber doesn't have an exact equivalent to gin.SetMode. It relies on the Config.

	r := fiber.New(fiber.Config{
		DisableStartupMessage: cfg.AppEnv == "production",
	})
	r.Use(recover.New())

	// ── Global middleware ──────────────────────────────────────────────
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))
	r.Use(middleware.LoggerMiddleware(log))

	// ── JWT Manager ────────────────────────────────────────────────────
	jwtManager := jwt.New(cfg.JWTSecret, cfg.JWTExpiredHours)

	// ── Health check ───────────────────────────────────────────────────
	r.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"status": "ok", "service": cfg.AppName})
	})

	// ── Swagger docs (non-production only) ────────────────────────────
	if cfg.AppEnv != "production" {
		r.Get("/swagger/*", swagger.HandlerDefault)
	}

	// ── Static file serving for uploads ───────────────────────────────
	r.Static("/uploads", cfg.StorageLocalPath)

	// ── API v1 routes ─────────────────────────────────────────────────
	api := r.Group("/api/v1")
	v1.Register(api, db, minioStorage, cryptoSvc, mailer, producer, osClient, jwtManager, log)

	return r
}

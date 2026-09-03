package router

import (
	"time"

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
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	// Handler imports (will be wired during full implementation)
	v1 "github.com/monitoring-system/backend/router/v1"
)

// Setup initialises Fiber, applies global middleware, and registers all API routes.
func Setup(cfg *config.Config, db *gorm.DB, redisClient *redis.Client, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, mailer mail.Mailer, producer kafka.EventProducer, osClient *opensearch.Client, log *logger.Logger) *fiber.App {
	// Fiber doesn't have an exact equivalent to gin.SetMode. It relies on the Config.

	r := fiber.New(fiber.Config{
		DisableStartupMessage: cfg.AppEnv == "production",
		BodyLimit:             50 * 1024 * 1024, // 50MB max upload body limit
		// Timeout settings — PENTING untuk mencegah koneksi hang di high concurrency (200 VU)
		// Tanpa ini, request yang stuck di DB connection pool wait bisa hang selamanya
		ReadTimeout:  30 * time.Second, // max waktu baca request dari client
		WriteTimeout: 30 * time.Second, // max waktu kirim response ke client
		IdleTimeout:  65 * time.Second, // keep-alive connection max idle time
	})
	r.Use(recover.New())

	// ── Global middleware ──────────────────────────────────────────────
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))
	// Applied globally (not per route-group) because every route file wires
	// its own AuthMiddleware instance — CSRFMiddleware itself no-ops for
	// GET/HEAD/OPTIONS and for requests without a session cookie, so it's
	// safe to run ahead of routing for every module.
	r.Use(middleware.CSRFMiddleware())
	r.Use(middleware.LoggerMiddleware(log))

	// Disable HTTP caching for all API responses to prevent browser cache stale data
	r.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Next()
	})

	// ── JWT Manager ────────────────────────────────────────────────────
	jwtManager := jwt.New(cfg.JWTSecret, cfg.AccessTokenTTL)

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
	// Baseline rate limit for every API endpoint — previously only
	// login/refresh/forgot-password/reset-password and the public KPI share
	// route had any limit at all, leaving everything else (including
	// self-service change-password, which accepts a guessable secret) open
	// to unlimited brute-force/enumeration/scraping. Endpoints that need a
	// tighter limit (login, refresh, password reset) layer their own
	// stricter AuthRateLimiter() on top of this.
	api.Use(middleware.APIRateLimiter())
	v1.Register(r, api, db, redisClient, minioStorage, cryptoSvc, mailer, producer, osClient, jwtManager, log, cfg)

	return r
}

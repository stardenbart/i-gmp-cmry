package router

import (
	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/opensearch"
	"github.com/monitoring-system/backend/pkg/storage"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	// Handler imports (will be wired during full implementation)
	v1 "github.com/monitoring-system/backend/router/v1"
)

// Setup initialises Gin, applies global middleware, and registers all API routes.
func Setup(cfg *config.Config, db *gorm.DB, minioStorage *storage.MinioStorage, cryptoSvc *crypto.Service, mailer mail.Mailer, producer kafka.EventProducer, osClient *opensearch.Client, log *logger.Logger) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// ── Global middleware ──────────────────────────────────────────────
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))
	r.Use(middleware.LoggerMiddleware(log))

	// ── JWT Manager ────────────────────────────────────────────────────
	jwtManager := jwt.New(cfg.JWTSecret, cfg.JWTExpiredHours)

	// ── Health check ───────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": cfg.AppName})
	})

	// ── Swagger docs (non-production only) ────────────────────────────
	if cfg.AppEnv != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// ── Static file serving for uploads ───────────────────────────────
	r.Static("/uploads", cfg.StorageLocalPath)

	// ── API v1 routes ─────────────────────────────────────────────────
	api := r.Group("/api/v1")
	v1.Register(api, db, minioStorage, cryptoSvc, mailer, producer, osClient, jwtManager, log)

	return r
}

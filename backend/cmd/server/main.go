package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	// @title           Monitoring Audit API
	// @version         1.0
	// @description     API untuk sistem monitoring audit internal perusahaan
	// @termsOfService  http://swagger.io/terms/

	// @contact.name   Internal Dev Team
	// @contact.email  dev@company.com

	// @license.name  MIT

	// @host      localhost:8080
	// @BasePath  /api/v1

	// @securityDefinitions.apikey BearerAuth
	// @in header
	// @name Authorization
	// @description Type "Bearer" followed by a space and JWT token.

	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/internal/domain/events"
	kafkainfra "github.com/monitoring-system/backend/internal/infrastructure/kafka"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/loggingrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/uploadrepo"
	"github.com/monitoring-system/backend/internal/usecase/uploadusecase"
	"github.com/monitoring-system/backend/internal/worker"
	"github.com/monitoring-system/backend/pkg/crypto"
	pkgkafka "github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/opensearch"
	"github.com/monitoring-system/backend/pkg/storage"
	"github.com/monitoring-system/backend/router"
)

func main() {
	// ── Load config & env ──────────────────────────────────────────────
	cfg := config.Load()

	// ── Initialize logger ──────────────────────────────────────────────
	log := logger.New(cfg.LogLevel, cfg.LogOutput, cfg.LogFilePath)
	defer log.Sync() //nolint:errcheck

	// ── Connect database ───────────────────────────────────────────────
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatal("failed to connect database", logger.Error(err))
	}

	// ── Prepare ActivityLog repo (still used for DB writes from Consumer) ──
	actLogRepo := loggingrepo.NewActivityLogRepository(db)
	_ = actLogRepo // used by OpenSearch indexer worker via direct DB writes (optional future path)

	// ── Setup SMTP Mailer ───────────────────────────────────────────────
	mailer := mail.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPSenderEmail)

	// ── Setup Kafka Producer ───────────────────────────────────────────
	kafkaBrokers := []string{cfg.KafkaBrokers}
	eventProducer := pkgkafka.NewProducer(kafkaBrokers)
	defer eventProducer.Close()

	// ── Setup OpenSearch ───────────────────────────────────────────────
	osClient, err := opensearch.NewClient(cfg.OpenSearchURL, cfg.OpenSearchUsername, cfg.OpenSearchPassword)
	if err != nil {
		log.Fatal("failed to connect opensearch", logger.Error(err))
	}

	// ── Setup Kafka Consumers (Background Workers) ─────────────────────
	// Use a background context; it will be cancelled on graceful shutdown via defer.
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	osIndexer := worker.NewOpenSearchIndexer(osClient, log)

	consumerInspections := pkgkafka.NewConsumer(kafkaBrokers, cfg.KafkaConsumerGroup, events.TopicAuditInspections, log)
	go consumerInspections.Start(bgCtx, osIndexer.HandleInspectionEvent)

	consumerIssues := pkgkafka.NewConsumer(kafkaBrokers, cfg.KafkaConsumerGroup, events.TopicAuditIssues, log)
	go consumerIssues.Start(bgCtx, osIndexer.HandleIssueEvent)

	consumerActivityLogs := pkgkafka.NewConsumer(kafkaBrokers, cfg.KafkaConsumerGroup, events.TopicActivityLogs, log)
	go consumerActivityLogs.Start(bgCtx, osIndexer.HandleActivityLogEvent)

	// ── Setup MinIO ────────────────────────────────────────────────────
	minioStorage, err := storage.NewMinioStorage(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioAllowedIPs, cfg.MinioUseSSL)
	if err != nil {
		log.Fatal("failed to connect minio", logger.Error(err))
	}

	// ── Sync MinIO IP Whitelist from DB on startup ─────────────────────
	settingRepo := masterrepo.NewSettingRepository(db)
	if s, err := settingRepo.FindByKey("MINIO_ALLOWED_IPS"); err == nil && s.SettingValue != "" {
		_ = minioStorage.UpdateIPWhitelistPolicy(context.Background(), s.SettingValue)
		log.Info("MinIO IP whitelist synced from DB")
	}

	// ── Init Crypto Service ────────────────────────────────────────────
	var cryptoSvc *crypto.Service
	if cfg.SettingEncryptionKey != "" {
		cryptoSvc, err = crypto.New(cfg.SettingEncryptionKey)
		if err != nil {
			log.Fatal("failed to init encryption service", logger.Error(err))
		}
	} else {
		log.Info("WARNING: SETTING_ENCRYPTION_KEY not set, encrypted settings will not be protected")
	}

	// ── Setup Image Processing Consumer ────────────────────────────────
	uploadRepo := uploadrepo.NewUploadRepository(db)
	imageProc := uploadusecase.NewImageProcessor(eventProducer)

	imageProcessingConsumer := kafkainfra.NewImageProcessingConsumer(
		kafkaBrokers,
		cfg.KafkaConsumerGroup,
		"audit.image-processing",
		minioStorage,
		imageProc,
		uploadRepo,
		log,
	)
	go imageProcessingConsumer.Start(bgCtx)

	// ── Setup router ───────────────────────────────────────────────────
	r := router.Setup(cfg, db, minioStorage, cryptoSvc, mailer, eventProducer, osClient, log)

	// ── HTTP Server ────────────────────────────────────────────────────
	// ── Graceful shutdown ──────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server starting", logger.String("port", cfg.AppPort))
		if err := r.Listen(":" + cfg.AppPort); err != nil {
			log.Fatal("server error", logger.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down server...")

	// Cancel all Kafka consumers before shutdown
	bgCancel()

	if err := r.Shutdown(); err != nil {
		log.Fatal("server forced to shutdown", logger.Error(err))
	}

	log.Info("server exited gracefully")
}

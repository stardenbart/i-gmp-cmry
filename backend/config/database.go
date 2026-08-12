package config

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates and validates a GORM database connection using PostgreSQL.
func NewDatabase(cfg *Config) (*gorm.DB, error) {
	// DSN format: host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Jakarta
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
		cfg.DBTimezone,
	)

	gormCfg := &gorm.Config{
		PrepareStmt: true,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// Show SQL queries only in development
	if cfg.AppEnv == "development" {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormCfg.Logger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Ensure high-performance indexes exist
	_ = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_userid ON "Users" ("UserID");
		CREATE INDEX IF NOT EXISTS idx_users_username ON "Users" ("Username");
		CREATE INDEX IF NOT EXISTS idx_activitylog_createdat ON "Activity_Log" ("ActivityCreatedAt" DESC);
		CREATE INDEX IF NOT EXISTS idx_activitylog_userid ON "Activity_Log" ("UserID");
		CREATE INDEX IF NOT EXISTS idx_hei_category_code ON "HEI_Master" ("CategoryName", "HEICode");
		CREATE INDEX IF NOT EXISTS idx_pic_mapping_userid ON "PIC_Mapping" ("UserID");
		CREATE INDEX IF NOT EXISTS idx_kawasan_aspek_kawasanid ON "kawasan_aspek" ("KawasanID");
	`)

	return db, nil
}

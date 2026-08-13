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

	// Ensure high-performance indexes and uploads table exist
	_ = db.Exec(`
		CREATE TABLE IF NOT EXISTS "uploads" (
			"id"                VARCHAR(50)  NOT NULL,
			"inspection_id"     VARCHAR(50)  NOT NULL,
			"original_filename" VARCHAR(255) NOT NULL,
			"stored_filename"   VARCHAR(255) NOT NULL,
			"file_path"         TEXT         NOT NULL,
			"file_size"         BIGINT       NOT NULL,
			"content_type"      VARCHAR(100) NOT NULL,
			"file_type"         VARCHAR(20)  NOT NULL,
			"status"            VARCHAR(20)  NOT NULL DEFAULT 'completed',
			"processed_url"     TEXT,
			"created_at"        BIGINT       NOT NULL,
			"updated_at"        BIGINT       NOT NULL,
			PRIMARY KEY ("id")
		);
		ALTER TABLE IF EXISTS "uploads" ALTER COLUMN "inspection_id" TYPE VARCHAR(50);
		ALTER TABLE IF EXISTS "uploads" ALTER COLUMN "id" TYPE VARCHAR(50);
		CREATE INDEX IF NOT EXISTS idx_uploads_inspection_id ON "uploads" ("inspection_id");
		CREATE INDEX IF NOT EXISTS idx_uploads_status ON "uploads" ("status");
		CREATE INDEX IF NOT EXISTS idx_users_userid ON "Users" ("UserID");
		-- Hapus constraint lama yang dicoba drop oleh GORM AutoMigrate (constraint mungkin tidak ada)
		-- Gunakan IF EXISTS agar tidak error saat restart
		ALTER TABLE IF EXISTS "Users" DROP CONSTRAINT IF EXISTS "uni_Users_username";
		ALTER TABLE IF EXISTS "Users" DROP CONSTRAINT IF EXISTS "uni_Users_email";
		-- Pastikan unique index versi kita tetap ada
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_unique ON "Users" ("Username");
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON "Users" ("Email");
		CREATE INDEX IF NOT EXISTS idx_activitylog_createdat ON "Activity_Log" ("ActivityCreatedAt" DESC);
		CREATE INDEX IF NOT EXISTS idx_activitylog_userid ON "Activity_Log" ("UserID");
		CREATE INDEX IF NOT EXISTS idx_hei_category_code ON "HEI_Master" ("CategoryName", "HEICode");
		CREATE INDEX IF NOT EXISTS idx_pic_mapping_userid ON "PIC_Mapping" ("UserID");
		CREATE INDEX IF NOT EXISTS idx_pic_mapping_area_kawasan ON "PIC_Mapping" ("AreaID", "KawasanID");
		CREATE INDEX IF NOT EXISTS idx_kawasan_aspek_kawasanid ON "kawasan_aspek" ("KawasanID");
		CREATE INDEX IF NOT EXISTS idx_dept_master_id ON "Department_Master" ("DepartmentID");
		CREATE INDEX IF NOT EXISTS idx_role_master_id ON "Role_Master" ("RoleID");
		CREATE INDEX IF NOT EXISTS idx_role_perm_role_isallowed ON "Role_Permission" ("RoleID", "IsAllowed");
		CREATE INDEX IF NOT EXISTS idx_user_perm_userid ON "User_Permission" ("UserID");
		CREATE INDEX IF NOT EXISTS idx_perm_master_mod_code ON "Permission_Master" ("ModuleID", "PermissionCode");
		CREATE INDEX IF NOT EXISTS idx_inspection_result_inspectionid ON "Inspection_Result" ("InspectionID");
		CREATE INDEX IF NOT EXISTS idx_inspection_result_id_checking ON "Inspection_Result" ("InspectionID", "Checking");
		CREATE INDEX IF NOT EXISTS idx_inspection_header_inspectionid ON "Inspection_Header" ("InspectionID");
		CREATE INDEX IF NOT EXISTS idx_insp_hdr_area_status ON "Inspection_Header" ("AreaID", "InspectionHeaderStatus");
		CREATE INDEX IF NOT EXISTS idx_insp_hdr_kawasan_status ON "Inspection_Header" ("KawasanID", "InspectionHeaderStatus");
		CREATE INDEX IF NOT EXISTS idx_insp_hdr_inspector_status ON "Inspection_Header" ("InspectorID", "InspectionHeaderStatus");
		CREATE INDEX IF NOT EXISTS idx_insp_hdr_created_at ON "Inspection_Header" ("InspectionHeaderCreatedAt" DESC);
		CREATE INDEX IF NOT EXISTS idx_area_master_plant ON "Area_Master" ("PlantID", "AreaID");
		CREATE INDEX IF NOT EXISTS idx_issue_photo_issueid ON "Issue_Photo" ("IssueID");
		CREATE INDEX IF NOT EXISTS idx_issue_photo_issueid_createdat ON "Issue_Photo" ("IssueID", "PhotoCreatedAt" ASC);
		CREATE INDEX IF NOT EXISTS idx_issue_resultid ON "Issue" ("ResultID");
		CREATE INDEX IF NOT EXISTS idx_issue_createdat ON "Issue" ("IssueCreatedAt" DESC);
		CREATE INDEX IF NOT EXISTS idx_notification_userid ON "Notification" ("UserID");
		CREATE INDEX IF NOT EXISTS idx_notification_userid_createdat ON "Notification" ("UserID", "CreatedAt" DESC);

		-- Index untuk mempercepat slow SQL filter inspection (IN subquery + status + created_at)
		-- Sebelumnya: 300-440ms. Target: <50ms dengan covering index
		CREATE INDEX IF NOT EXISTS idx_insp_hdr_area_status_created ON "Inspection_Header" ("AreaID", "InspectionHeaderStatus", "InspectionHeaderCreatedAt" DESC);
		CREATE INDEX IF NOT EXISTS idx_insp_hdr_inspector_created ON "Inspection_Header" ("InspectorID", "InspectionHeaderCreatedAt" DESC);
		-- Covering index untuk subquery Area_Master: mempercepat IN (SELECT AreaID FROM Area_Master WHERE PlantID = ...)
		CREATE INDEX IF NOT EXISTS idx_area_master_plant_areaid ON "Area_Master" ("PlantID", "AreaID") INCLUDE ("AreaName");
		-- Activity_Log: index tambahan untuk filter by module + action
		CREATE INDEX IF NOT EXISTS idx_activitylog_table_action ON "Activity_Log" ("TableAffected", "ActivityAction", "ActivityCreatedAt" DESC);
	`)

	return db, nil
}

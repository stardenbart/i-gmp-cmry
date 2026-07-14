-- =========================================================
-- Migration 007: System Settings Table
-- Stores dynamic application configuration as key-value pairs.
-- =========================================================

CREATE TABLE IF NOT EXISTS "System_Setting" (
    "SettingKey"   VARCHAR(100) NOT NULL,
    "SettingValue" TEXT         NOT NULL,
    "IsEncrypted"  BOOLEAN      NOT NULL DEFAULT FALSE,
    "Description"  VARCHAR(255),
    "UpdatedAt"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedBy"    VARCHAR(20),
    PRIMARY KEY ("SettingKey"),
    CONSTRAINT "fk_setting_updatedby" FOREIGN KEY ("UpdatedBy") REFERENCES "Users" ("UserID")
);

-- Insert default values
INSERT INTO "System_Setting" ("SettingKey", "SettingValue", "IsEncrypted", "Description")
VALUES
    ('MINIO_ALLOWED_IPS', '127.0.0.1/32,10.0.0.0/8,192.168.0.0/16,172.16.0.0/12', FALSE, 'Comma-separated CIDR ranges allowed to access MinIO S3 bucket'),
    ('MAX_UPLOAD_SIZE_MB', '10', FALSE, 'Maximum file upload size in MB'),
    ('MAX_LOGIN_ATTEMPTS', '5', FALSE, 'Maximum failed login attempts before account lockout'),
    ('SESSION_IDLE_TIMEOUT_MINUTES', '60', FALSE, 'User session idle timeout in minutes')
ON CONFLICT ("SettingKey") DO NOTHING;

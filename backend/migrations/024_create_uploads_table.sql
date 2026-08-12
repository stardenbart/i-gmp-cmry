-- Migration 024: Create uploads table for file storage tracking
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

-- =========================================================
-- Migration 022: Add HEICategory column to Issue_Photo table
-- =========================================================

ALTER TABLE "Issue_Photo" ADD COLUMN IF NOT EXISTS "HEICategory" VARCHAR(50);
CREATE INDEX IF NOT EXISTS "idx_issue_photo_hei_category" ON "Issue_Photo" ("HEICategory");

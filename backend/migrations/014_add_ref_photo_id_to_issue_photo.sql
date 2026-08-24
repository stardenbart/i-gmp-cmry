-- =========================================================
-- Migration 014: Add RefPhotoID to Issue_Photo table
-- =========================================================

ALTER TABLE "Issue_Photo" ADD COLUMN IF NOT EXISTS "RefPhotoID" VARCHAR(50);

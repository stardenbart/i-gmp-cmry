-- =========================================================
-- Migration 008: Alter Tables for New Features
-- =========================================================

ALTER TABLE "Inspection_Header" 
ADD COLUMN IF NOT EXISTS "SessionID" VARCHAR(255),
ADD COLUMN IF NOT EXISTS "LockedAt" TIMESTAMP;

ALTER TABLE "Issue"
ADD COLUMN IF NOT EXISTS "Label" VARCHAR(100) DEFAULT '',
ADD COLUMN IF NOT EXISTS "NeedsWOWR" BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS "WO_ID" VARCHAR(100) DEFAULT '',
ADD COLUMN IF NOT EXISTS "WR_ID" VARCHAR(100) DEFAULT '';

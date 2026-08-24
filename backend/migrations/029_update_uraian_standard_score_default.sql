-- =========================================================
-- Migration 029: Update default StandardScore in Uraian_Master to 2
-- =========================================================

ALTER TABLE "Uraian_Master" ALTER COLUMN "StandardScore" SET DEFAULT 2;

-- Update existing Uraian records with StandardScore = 0 or 100 to 2
UPDATE "Uraian_Master" SET "StandardScore" = 2 WHERE "StandardScore" = 0 OR "StandardScore" = 100;

-- =========================================================
-- Migration 019: Add PlantID to Users table for Per-Plant Authorization
-- =========================================================

ALTER TABLE "Users" ADD COLUMN IF NOT EXISTS "PlantID" VARCHAR(20);

-- Nullify any invalid PlantID values on Users table before adding FK constraint
UPDATE "Users"
SET "PlantID" = NULL
WHERE "PlantID" IS NOT NULL AND "PlantID" NOT IN (SELECT "PlantID" FROM "Plant_Master");

ALTER TABLE "Users" DROP CONSTRAINT IF EXISTS "fk_users_plant";
ALTER TABLE "Users" ADD CONSTRAINT "fk_users_plant" FOREIGN KEY ("PlantID") REFERENCES "Plant_Master" ("PlantID") ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS "idx_users_plant" ON "Users" ("PlantID");

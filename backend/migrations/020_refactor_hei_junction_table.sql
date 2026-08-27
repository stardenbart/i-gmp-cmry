-- =========================================================
-- Migration 020: Refactor HEI to Issue_HEI Junction Table (1-to-1 per Issue)
-- =========================================================

-- 1. Create Table Issue_HEI
CREATE TABLE IF NOT EXISTS "Issue_HEI" (
    "IssueHEIID"         VARCHAR(50)  NOT NULL,
    "IssueID"            VARCHAR(50)  NOT NULL,
    "HabitID"            VARCHAR(50),
    "EquipmentID"        VARCHAR(50),
    "InfrastructureID"   VARCHAR(20),
    "CreatedAt"          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt"          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("IssueHEIID"),
    CONSTRAINT "fk_issuehei_issue" FOREIGN KEY ("IssueID") REFERENCES "Issue" ("IssueID") ON DELETE CASCADE,
    CONSTRAINT "fk_issuehei_habit" FOREIGN KEY ("HabitID") REFERENCES "Habit_Master" ("HabitID") ON DELETE SET NULL,
    CONSTRAINT "fk_issuehei_equipment" FOREIGN KEY ("EquipmentID") REFERENCES "Equipment_Master" ("EquipmentID") ON DELETE SET NULL,
    CONSTRAINT "fk_issuehei_infra" FOREIGN KEY ("InfrastructureID") REFERENCES "Infrastructure_Master" ("InfrastructureID") ON DELETE SET NULL,
    CONSTRAINT "unique_issue_hei" UNIQUE ("IssueID"),
    CONSTRAINT "chk_issuehei_at_least_one" CHECK (
        "HabitID" IS NOT NULL OR "EquipmentID" IS NOT NULL OR "InfrastructureID" IS NOT NULL
    )
);

-- The application repository may have created Issue_HEI through AutoMigrate
-- before the SQL migration runner starts. In that case the table already uses
-- the newer unified HEIID shape, so CREATE TABLE IF NOT EXISTS above is a
-- no-op. Add the legacy transition columns explicitly so this migration can
-- migrate Issue data and migration 028 can consolidate it back to HEIID.
ALTER TABLE "Issue_HEI" ADD COLUMN IF NOT EXISTS "HabitID" VARCHAR(50);
ALTER TABLE "Issue_HEI" ADD COLUMN IF NOT EXISTS "EquipmentID" VARCHAR(50);
ALTER TABLE "Issue_HEI" ADD COLUMN IF NOT EXISTS "InfrastructureID" VARCHAR(50);

CREATE INDEX IF NOT EXISTS "idx_issuehei_issue_id" ON "Issue_HEI" ("IssueID");
CREATE INDEX IF NOT EXISTS "idx_issuehei_habit" ON "Issue_HEI" ("HabitID") WHERE "HabitID" IS NOT NULL;
CREATE INDEX IF NOT EXISTS "idx_issuehei_equipment" ON "Issue_HEI" ("EquipmentID") WHERE "EquipmentID" IS NOT NULL;
CREATE INDEX IF NOT EXISTS "idx_issuehei_infrastructure" ON "Issue_HEI" ("InfrastructureID") WHERE "InfrastructureID" IS NOT NULL;

-- 2. Migrate existing data from Issue table to Issue_HEI table
INSERT INTO "Issue_HEI" ("IssueHEIID", "IssueID", "HabitID", "EquipmentID", "InfrastructureID", "CreatedAt", "UpdatedAt")
SELECT 
    'HEI-' || "IssueID",
    "IssueID",
    "HabitID",
    "EquipmentID",
    "InfrastructureID",
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM "Issue"
WHERE "HabitID" IS NOT NULL OR "EquipmentID" IS NOT NULL OR "InfrastructureID" IS NOT NULL
ON CONFLICT ("IssueID") DO UPDATE SET
    "HabitID" = EXCLUDED."HabitID",
    "EquipmentID" = EXCLUDED."EquipmentID",
    "InfrastructureID" = EXCLUDED."InfrastructureID",
    "UpdatedAt" = CURRENT_TIMESTAMP;

-- 3. Drop old CHECK constraint and columns from Issue table
ALTER TABLE "Issue" DROP CONSTRAINT IF EXISTS "chk_issue_category";
DROP INDEX IF EXISTS "idx_issue_habit";
DROP INDEX IF EXISTS "idx_issue_equipment";
DROP INDEX IF EXISTS "idx_issue_infrastructure";

ALTER TABLE "Issue" DROP COLUMN IF EXISTS "IssueCategory";
ALTER TABLE "Issue" DROP COLUMN IF EXISTS "HabitID";
ALTER TABLE "Issue" DROP COLUMN IF EXISTS "EquipmentID";
ALTER TABLE "Issue" DROP COLUMN IF EXISTS "InfrastructureID";

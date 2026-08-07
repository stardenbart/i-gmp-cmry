-- =========================================================
-- Migration 020 Rollback: Restore HEI columns to Issue table
-- =========================================================

-- 1. Restore columns in Issue table
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "IssueCategory" issue_category_enum;
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "HabitID" VARCHAR(20) REFERENCES "Habit_Master" ("HabitID") ON DELETE RESTRICT;
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "EquipmentID" VARCHAR(20) REFERENCES "Equipment_Master" ("EquipmentID") ON DELETE RESTRICT;
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "InfrastructureID" VARCHAR(20) REFERENCES "Infrastructure_Master" ("InfrastructureID") ON DELETE RESTRICT;

-- 2. Restore data from Issue_HEI to Issue table
UPDATE "Issue" i
SET 
    "HabitID" = ih."HabitID",
    "EquipmentID" = ih."EquipmentID",
    "InfrastructureID" = ih."InfrastructureID",
    "IssueCategory" = CASE 
        WHEN ih."HabitID" IS NOT NULL THEN 'Habit'::issue_category_enum
        WHEN ih."EquipmentID" IS NOT NULL THEN 'Equipment'::issue_category_enum
        WHEN ih."InfrastructureID" IS NOT NULL THEN 'Infrastructure'::issue_category_enum
        ELSE NULL
    END
FROM "Issue_HEI" ih
WHERE i."IssueID" = ih."IssueID";

-- 3. Restore indexes and constraint
CREATE INDEX IF NOT EXISTS "idx_issue_habit" ON "Issue" ("HabitID") WHERE "HabitID" IS NOT NULL;
CREATE INDEX IF NOT EXISTS "idx_issue_equipment" ON "Issue" ("EquipmentID") WHERE "EquipmentID" IS NOT NULL;
CREATE INDEX IF NOT EXISTS "idx_issue_infrastructure" ON "Issue" ("InfrastructureID") WHERE "InfrastructureID" IS NOT NULL;

ALTER TABLE "Issue" DROP CONSTRAINT IF EXISTS "chk_issue_category";
ALTER TABLE "Issue" ADD CONSTRAINT "chk_issue_category" CHECK (
    ("IssueCategory" = 'Habit'          AND "HabitID" IS NOT NULL AND "EquipmentID" IS NULL AND "InfrastructureID" IS NULL) OR
    ("IssueCategory" = 'Equipment'      AND "EquipmentID" IS NOT NULL AND "HabitID" IS NULL AND "InfrastructureID" IS NULL) OR
    ("IssueCategory" = 'Infrastructure' AND "InfrastructureID" IS NOT NULL AND "HabitID" IS NULL AND "EquipmentID" IS NULL) OR
    ("IssueCategory" IS NULL AND "HabitID" IS NULL AND "EquipmentID" IS NULL AND "InfrastructureID" IS NULL)
);

-- 4. Drop Issue_HEI table
DROP TABLE IF EXISTS "Issue_HEI";

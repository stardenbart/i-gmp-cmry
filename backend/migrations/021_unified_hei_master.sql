-- =========================================================
-- Migration 021: Unified HEI Master Table & Dynamic Categories
-- =========================================================

-- 1. Create HEI_Master table
CREATE TABLE IF NOT EXISTS "HEI_Master" (
    "HEIID"        VARCHAR(50)  NOT NULL,
    "CategoryName" VARCHAR(50)  NOT NULL,
    "HEICode"      VARCHAR(50),
    "HEIName"      VARCHAR(150) NOT NULL,
    "Description"  VARCHAR(255),
    "Status"       VARCHAR(20)  NOT NULL DEFAULT 'Active',
    "CreatedAt"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("HEIID")
);

CREATE INDEX IF NOT EXISTS "idx_hei_category" ON "HEI_Master" ("CategoryName");
CREATE INDEX IF NOT EXISTS "idx_hei_code" ON "HEI_Master" ("HEICode");

-- 2. Migrate existing Habit_Master to HEI_Master
INSERT INTO "HEI_Master" ("HEIID", "CategoryName", "HEICode", "HEIName", "Description", "Status", "CreatedAt", "UpdatedAt")
SELECT 
    "HabitID",
    'Habit',
    "HabitCode",
    "HabitName",
    "Description",
    'Active',
    "HabitCreatedAt",
    "HabitUpdatedAt"
FROM "Habit_Master"
ON CONFLICT ("HEIID") DO NOTHING;

-- 3. Migrate existing Equipment_Master to HEI_Master
INSERT INTO "HEI_Master" ("HEIID", "CategoryName", "HEICode", "HEIName", "Description", "Status", "CreatedAt", "UpdatedAt")
SELECT 
    "EquipmentID",
    'Equipment',
    "EquipmentCode",
    "EquipmentName",
    "EquipmentType",
    "EquipmentStatus",
    "EquipmentCreatedAt",
    "EquipmentUpdatedAt"
FROM "Equipment_Master"
ON CONFLICT ("HEIID") DO NOTHING;

-- 4. Migrate existing Infrastructure_Master to HEI_Master
INSERT INTO "HEI_Master" ("HEIID", "CategoryName", "HEICode", "HEIName", "Description", "Status", "CreatedAt", "UpdatedAt")
SELECT 
    "InfrastructureID",
    'Infrastructure',
    "InfrastructureCode",
    "InfrastructureName",
    "InfrastructureType",
    "InfrastructureStatus",
    "InfrastructureCreatedAt",
    "InfrastructureUpdatedAt"
FROM "Infrastructure_Master"
ON CONFLICT ("HEIID") DO NOTHING;

-- 5. Add HEIID column to Issue_Photo
ALTER TABLE "Issue_Photo" ADD COLUMN IF NOT EXISTS "HEIID" VARCHAR(50) REFERENCES "HEI_Master" ("HEIID") ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS "idx_issue_photo_hei_id" ON "Issue_Photo" ("HEIID");

-- 6. Populate HEIID from HabitID, EquipmentID, or InfrastructureID
UPDATE "Issue_Photo"
SET "HEIID" = COALESCE("HabitID", "EquipmentID", "InfrastructureID")
WHERE "HEIID" IS NULL AND ("HabitID" IS NOT NULL OR "EquipmentID" IS NOT NULL OR "InfrastructureID" IS NOT NULL);

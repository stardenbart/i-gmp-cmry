-- =========================================================
-- Migration 018: ERD v2 Schema Upgrade (Plant, HEI Framework & Fixes)
-- =========================================================

-- 1. Create Enum issue_category_enum
CREATE TYPE issue_category_enum AS ENUM ('Habit', 'Equipment', 'Infrastructure');

-- 2. Create Table Plant_Master
CREATE TABLE IF NOT EXISTS "Plant_Master" (
    "PlantID"        VARCHAR(20)  NOT NULL,
    "PlantCode"      VARCHAR(20)  NOT NULL,
    "PlantName"      VARCHAR(100) NOT NULL,
    "Address"        VARCHAR(255),
    "PlantCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "PlantUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("PlantID"),
    CONSTRAINT "unique_plant_code" UNIQUE ("PlantCode")
);

-- 3. Create Table Habit_Master
CREATE TABLE IF NOT EXISTS "Habit_Master" (
    "HabitID"        VARCHAR(20)  NOT NULL,
    "HabitCode"      VARCHAR(20),
    "HabitName"      VARCHAR(150) NOT NULL,
    "HabitCategory"  VARCHAR(50),
    "Description"    VARCHAR(255),
    "HabitCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "HabitUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("HabitID"),
    CONSTRAINT "unique_habit_code" UNIQUE ("HabitCode")
);

-- 4. Create Table Equipment_Master
CREATE TABLE IF NOT EXISTS "Equipment_Master" (
    "EquipmentID"        VARCHAR(20)  NOT NULL,
    "KawasanID"          VARCHAR(20)  NOT NULL,
    "EquipmentCode"      VARCHAR(50),
    "EquipmentName"      VARCHAR(150) NOT NULL,
    "EquipmentType"      VARCHAR(50),
    "EquipmentStatus"    VARCHAR(20)  NOT NULL DEFAULT 'Active',
    "EquipmentCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "EquipmentUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("EquipmentID"),
    CONSTRAINT "fk_equipment_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID") ON DELETE RESTRICT,
    CONSTRAINT "unique_equipment_code" UNIQUE ("EquipmentCode")
);
CREATE INDEX IF NOT EXISTS "idx_equipment_kawasan" ON "Equipment_Master" ("KawasanID");

-- 5. Create Table Infrastructure_Master
CREATE TABLE IF NOT EXISTS "Infrastructure_Master" (
    "InfrastructureID"        VARCHAR(20)  NOT NULL,
    "KawasanID"               VARCHAR(20)  NOT NULL,
    "InfrastructureCode"      VARCHAR(50),
    "InfrastructureName"      VARCHAR(150) NOT NULL,
    "InfrastructureType"      VARCHAR(50),
    "InfrastructureStatus"    VARCHAR(20)  NOT NULL DEFAULT 'Active',
    "InfrastructureCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "InfrastructureUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("InfrastructureID"),
    CONSTRAINT "fk_infra_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID") ON DELETE RESTRICT,
    CONSTRAINT "unique_infra_code" UNIQUE ("InfrastructureCode")
);
CREATE INDEX IF NOT EXISTS "idx_infrastructure_kawasan" ON "Infrastructure_Master" ("KawasanID");

-- 6. Alter Area_Master (add PlantID)
ALTER TABLE "Area_Master" ADD COLUMN IF NOT EXISTS "PlantID" VARCHAR(20) REFERENCES "Plant_Master" ("PlantID") ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS "idx_area_plant" ON "Area_Master" ("PlantID");

-- 7. Fix Inspection_Header column typo
ALTER TABLE "Inspection_Header" RENAME COLUMN "InspectionheaderUpdatedAt" TO "InspectionHeaderUpdatedAt";

-- 8. Fix PIC_Mapping AreaID default
ALTER TABLE "PIC_Mapping" ALTER COLUMN "AreaID" DROP DEFAULT;

-- 9. Fix Notification UserID type & FK constraint
ALTER TABLE "Notification" ALTER COLUMN "UserID" TYPE VARCHAR(20);
ALTER TABLE "Notification" DROP CONSTRAINT IF EXISTS "fk_notification_user";
ALTER TABLE "Notification" ADD CONSTRAINT "fk_notification_user" FOREIGN KEY ("UserID") REFERENCES "Users" ("UserID") ON DELETE CASCADE;

-- 10. Alter Issue (add HEI Category & foreign keys & CHECK constraint)
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "IssueCategory" issue_category_enum;
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "HabitID" VARCHAR(20) REFERENCES "Habit_Master" ("HabitID") ON DELETE RESTRICT;
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "EquipmentID" VARCHAR(20) REFERENCES "Equipment_Master" ("EquipmentID") ON DELETE RESTRICT;
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "InfrastructureID" VARCHAR(20) REFERENCES "Infrastructure_Master" ("InfrastructureID") ON DELETE RESTRICT;

ALTER TABLE "Issue" DROP CONSTRAINT IF EXISTS "chk_issue_category";
ALTER TABLE "Issue" ADD CONSTRAINT "chk_issue_category" CHECK (
    ("IssueCategory" = 'Habit'          AND "HabitID" IS NOT NULL AND "EquipmentID" IS NULL AND "InfrastructureID" IS NULL) OR
    ("IssueCategory" = 'Equipment'      AND "EquipmentID" IS NOT NULL AND "HabitID" IS NULL AND "InfrastructureID" IS NULL) OR
    ("IssueCategory" = 'Infrastructure' AND "InfrastructureID" IS NOT NULL AND "HabitID" IS NULL AND "EquipmentID" IS NULL) OR
    ("IssueCategory" IS NULL AND "HabitID" IS NULL AND "EquipmentID" IS NULL AND "InfrastructureID" IS NULL)
);

CREATE INDEX IF NOT EXISTS "idx_issue_habit" ON "Issue" ("HabitID") WHERE "HabitID" IS NOT NULL;
CREATE INDEX IF NOT EXISTS "idx_issue_equipment" ON "Issue" ("EquipmentID") WHERE "EquipmentID" IS NOT NULL;
CREATE INDEX IF NOT EXISTS "idx_issue_infrastructure" ON "Issue" ("InfrastructureID") WHERE "InfrastructureID" IS NOT NULL;

-- 11. Create index for inspeksi_audit_log SessionID
CREATE INDEX IF NOT EXISTS "idx_audit_session_id" ON "inspeksi_audit_log" ("SessionID");

-- =========================================================
-- Migration 028: Consolidate HEI Tables (Habit, Equipment, Infrastructure)
-- =========================================================

-- 1. Create unified HEI_Master table
CREATE TABLE IF NOT EXISTS "HEI_Master" (
    "HEIID"        VARCHAR(50)  NOT NULL,
    "CategoryName" VARCHAR(50)  NOT NULL, -- 'Habit', 'Equipment', 'Infrastructure'
    "HEICode"      VARCHAR(50),
    "HEIName"      VARCHAR(150) NOT NULL,
    "KawasanID"    VARCHAR(20),
    "Description"  VARCHAR(255),
    "Status"       VARCHAR(20)  NOT NULL DEFAULT 'Active',
    "CreatedAt"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("HEIID"),
    CONSTRAINT "fk_hei_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID") ON DELETE SET NULL
);

ALTER TABLE "HEI_Master" ADD COLUMN IF NOT EXISTS "KawasanID" VARCHAR(20) REFERENCES "Kawasan_Master" ("KawasanID") ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS "idx_hei_category" ON "HEI_Master" ("CategoryName");
CREATE INDEX IF NOT EXISTS "idx_hei_kawasan" ON "HEI_Master" ("KawasanID");

-- 2. Migrate existing Habit_Master data to HEI_Master
DO $$ BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'Habit_Master') THEN
        INSERT INTO "HEI_Master" ("HEIID", "CategoryName", "HEICode", "HEIName", "Description", "Status", "CreatedAt", "UpdatedAt")
        SELECT "HabitID", 'Habit', "HabitCode", "HabitName", "Description", 'Active', "HabitCreatedAt", "HabitUpdatedAt"
        FROM "Habit_Master"
        ON CONFLICT ("HEIID") DO NOTHING;
    END IF;
END $$;

-- 3. Migrate existing Equipment_Master data to HEI_Master
DO $$ BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'Equipment_Master') THEN
        INSERT INTO "HEI_Master" ("HEIID", "CategoryName", "HEICode", "HEIName", "KawasanID", "Description", "Status", "CreatedAt", "UpdatedAt")
        SELECT "EquipmentID", 'Equipment', "EquipmentCode", "EquipmentName", "KawasanID", "EquipmentType", "EquipmentStatus", "EquipmentCreatedAt", "EquipmentUpdatedAt"
        FROM "Equipment_Master"
        ON CONFLICT ("HEIID") DO NOTHING;
    END IF;
END $$;

-- 4. Migrate existing Infrastructure_Master data to HEI_Master
DO $$ BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'Infrastructure_Master') THEN
        INSERT INTO "HEI_Master" ("HEIID", "CategoryName", "HEICode", "HEIName", "KawasanID", "Description", "Status", "CreatedAt", "UpdatedAt")
        SELECT "InfrastructureID", 'Infrastructure', "InfrastructureCode", "InfrastructureName", "KawasanID", "InfrastructureType", "InfrastructureStatus", "InfrastructureCreatedAt", "InfrastructureUpdatedAt"
        FROM "Infrastructure_Master"
        ON CONFLICT ("HEIID") DO NOTHING;
    END IF;
END $$;

-- 5. Add HEIID column to Issue_HEI table
ALTER TABLE "Issue_HEI" ADD COLUMN IF NOT EXISTS "HEIID" VARCHAR(50) REFERENCES "HEI_Master" ("HEIID") ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS "idx_issuehei_hei_id" ON "Issue_HEI" ("HEIID");

-- 6. Populate HEIID from HabitID, EquipmentID, or InfrastructureID
DO $$ BEGIN
    IF EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'Issue_HEI' AND column_name = 'HabitID') THEN
        UPDATE "Issue_HEI"
        SET "HEIID" = COALESCE("HabitID", "EquipmentID", "InfrastructureID")
        WHERE "HEIID" IS NULL AND ("HabitID" IS NOT NULL OR "EquipmentID" IS NOT NULL OR "InfrastructureID" IS NOT NULL);
    END IF;
END $$;

-- 7. Drop old constraints and columns from Issue_HEI
ALTER TABLE "Issue_HEI" DROP CONSTRAINT IF EXISTS "chk_issuehei_at_least_one";
ALTER TABLE "Issue_HEI" DROP CONSTRAINT IF EXISTS "fk_issuehei_habit";
ALTER TABLE "Issue_HEI" DROP CONSTRAINT IF EXISTS "fk_issuehei_equipment";
ALTER TABLE "Issue_HEI" DROP CONSTRAINT IF EXISTS "fk_issuehei_infra";

ALTER TABLE "Issue_HEI" DROP COLUMN IF EXISTS "HabitID";
ALTER TABLE "Issue_HEI" DROP COLUMN IF EXISTS "EquipmentID";
ALTER TABLE "Issue_HEI" DROP COLUMN IF EXISTS "InfrastructureID";

-- 8. Safely drop obsolete master tables
DROP TABLE IF EXISTS "Habit_Master" CASCADE;
DROP TABLE IF EXISTS "Equipment_Master" CASCADE;
DROP TABLE IF EXISTS "Infrastructure_Master" CASCADE;

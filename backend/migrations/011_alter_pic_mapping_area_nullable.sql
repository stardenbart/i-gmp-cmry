-- =========================================================
-- Migration 011: Make AreaID nullable in PIC_Mapping
--                PIC mapping now works per Kawasan only
-- =========================================================

-- Drop the existing NOT NULL constraint and foreign key for AreaID
ALTER TABLE "PIC_Mapping" DROP CONSTRAINT IF EXISTS "fk_picmap_area";
ALTER TABLE "PIC_Mapping" ALTER COLUMN "AreaID" DROP NOT NULL;
ALTER TABLE "PIC_Mapping" ALTER COLUMN "AreaID" SET DEFAULT '';

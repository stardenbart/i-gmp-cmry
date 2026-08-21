-- Migration 027: Add DetailKawasanID column to Issue table
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "DetailKawasanID" VARCHAR(20);

-- Backfill DetailKawasanID for existing issues using Inspection_Result and Inspection_Header
UPDATE "Issue" i
SET "DetailKawasanID" = ih."DetailKawasanID"
FROM "Inspection_Result" ir
JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"
WHERE i."ResultID" = ir."ResultID"
  AND (i."DetailKawasanID" IS NULL OR i."DetailKawasanID" = '');

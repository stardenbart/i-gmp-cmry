-- Add LastInspection column to DetailKawasan_Master and Kawasan_Master
ALTER TABLE "DetailKawasan_Master" ADD COLUMN IF NOT EXISTS "LastInspection" TIMESTAMP;
ALTER TABLE "Kawasan_Master" ADD COLUMN IF NOT EXISTS "LastInspection" TIMESTAMP;

-- Add Keterangan column to Issue_Photo table for per-photo descriptions
ALTER TABLE "Issue_Photo" ADD COLUMN IF NOT EXISTS "Keterangan" VARCHAR(255);

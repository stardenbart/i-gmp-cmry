-- Add FollowUpDelay column to Issue table
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "FollowUpDelay" INTEGER;

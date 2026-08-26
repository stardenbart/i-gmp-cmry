-- Repair issues that were closed before photo-level WO/WR validation was
-- enforced. They must return to PendingValidation so an Admin/Auditor can
-- explicitly verify every referenced WO/WR before final closure.

-- These fields originally existed only on Issue and were later supported per
-- photo. Create them before the repair query so a fresh database can execute
-- the complete migration chain without relying on application AutoMigrate.
ALTER TABLE "Issue_Photo"
    ADD COLUMN IF NOT EXISTS "NeedsWOWR" BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS "WO_ID" VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "WR_ID" VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "WOWRStatus" VARCHAR(50) NOT NULL DEFAULT 'None';

UPDATE "Issue" AS i
SET
    "IssueStatus" = 'PendingValidation',
    "IssueUpdatedAt" = CURRENT_TIMESTAMP
WHERE i."IssueStatus" IN ('Closed', 'Verified')
  AND EXISTS (
      SELECT 1
      FROM "Issue_Photo" AS ip
      WHERE ip."IssueID" = i."IssueID"
        AND ip."PhotoType" = 'Initial'
        AND (
            ip."NeedsWOWR" = TRUE
            OR COALESCE(ip."WO_ID", '') <> ''
            OR COALESCE(ip."WR_ID", '') <> ''
        )
        AND COALESCE(ip."WOWRStatus", 'None') <> 'Verified'
  );

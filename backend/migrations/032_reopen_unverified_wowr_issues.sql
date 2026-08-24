-- Repair issues that were closed before photo-level WO/WR validation was
-- enforced. They must return to PendingValidation so an Admin/Auditor can
-- explicitly verify every referenced WO/WR before final closure.

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

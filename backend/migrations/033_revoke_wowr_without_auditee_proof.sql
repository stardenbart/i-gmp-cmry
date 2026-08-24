-- Revoke historical photo-level WO/WR approvals that do not have a linked
-- WOWR proof image uploaded by the Issue's assigned PIC/Auditee.

UPDATE "Issue_Photo" AS request_photo
SET
    "WOWRStatus" = 'PendingValidation',
    "PhotoUpdatedAt" = CURRENT_TIMESTAMP
FROM "Issue" AS i
WHERE i."IssueID" = request_photo."IssueID"
  AND request_photo."PhotoType" = 'Initial'
  AND request_photo."WOWRStatus" = 'Verified'
  AND (
      request_photo."NeedsWOWR" = TRUE
      OR COALESCE(request_photo."WO_ID", '') <> ''
      OR COALESCE(request_photo."WR_ID", '') <> ''
  )
  AND NOT EXISTS (
      SELECT 1
      FROM "Issue_Photo" AS proof
      WHERE proof."IssueID" = request_photo."IssueID"
        AND proof."PhotoType" = 'WOWR'
        AND proof."RefPhotoID" = request_photo."IssuePhotoID"
        AND proof."PICUserID" = i."IssuePICUserID"
        AND (
            COALESCE(proof."ImageUrl", '') <> ''
            OR COALESCE(proof."FileName", '') <> ''
        )
  );

UPDATE "Issue" AS i
SET
    "IssueStatus" = 'PendingValidation',
    "IssueUpdatedAt" = CURRENT_TIMESTAMP
WHERE i."IssueStatus" IN ('Closed', 'Verified')
  AND EXISTS (
      SELECT 1
      FROM "Issue_Photo" AS request_photo
      WHERE request_photo."IssueID" = i."IssueID"
        AND request_photo."PhotoType" = 'Initial'
        AND request_photo."WOWRStatus" = 'PendingValidation'
        AND (
            request_photo."NeedsWOWR" = TRUE
            OR COALESCE(request_photo."WO_ID", '') <> ''
            OR COALESCE(request_photo."WR_ID", '') <> ''
        )
  );

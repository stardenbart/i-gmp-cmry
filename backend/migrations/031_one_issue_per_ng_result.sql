-- Every NG inspection result must own an independent Issue. Previous versions
-- reused one active Issue for all findings in the same DetailKawasan, causing
-- photos from different uraian to overwrite or delete each other.

CREATE UNIQUE INDEX IF NOT EXISTS "idx_issue_resultid_unique"
    ON "Issue" ("ResultID");

INSERT INTO "Issue" (
    "IssueID",
    "ResultID",
    "DetailKawasanID",
    "IssuePICUserID",
    "DueDate",
    "IssueStatus",
    "Keterangan",
    "IssueCreatedAt",
    "IssueUpdatedAt"
)
SELECT
    'ISS-NG-' || LEFT(MD5(ir."ResultID"), 12),
    ir."ResultID",
    ih."DetailKawasanID",
    ih."InspectorID",
    CURRENT_DATE + 14,
    'Open',
    LEFT(
        COALESCE(
            NULLIF(ir."Keterangan", ''),
            'Temuan NG pada uraian ' || ir."UraianID"
        ),
        255
    ),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM "Inspection_Result" ir
JOIN "Inspection_Header" ih
  ON ih."InspectionID" = ir."InspectionID"
WHERE ir."Checking" = 'NG'
  AND ih."InspectionHeaderStatus" IN ('Completed', 'Approved')
  AND ih."InspectorID" IS NOT NULL
  AND ih."InspectorID" <> ''
  AND NOT EXISTS (
      SELECT 1
      FROM "Issue" existing_issue
      WHERE existing_issue."ResultID" = ir."ResultID"
  )
ON CONFLICT ("ResultID") DO NOTHING;

-- Backfill completed NG findings that were saved before issue synchronization
-- became a required part of the bulk-save operation.
--
-- The application maintains one active issue per DetailKawasan. DISTINCT ON
-- follows that invariant when several NG results exist in the same location.
WITH missing_ng AS (
    SELECT DISTINCT ON (ih."DetailKawasanID")
        ir."ResultID",
        ih."DetailKawasanID",
        ih."InspectorID",
        LEFT(
            COALESCE(
                NULLIF(ir."Keterangan", ''),
                'Temuan NG pada inspeksi ' || ih."InspectionID"
            ),
            255
        ) AS "Keterangan",
        ih."InspectionHeaderUpdatedAt"
    FROM "Inspection_Result" ir
    JOIN "Inspection_Header" ih
      ON ih."InspectionID" = ir."InspectionID"
    WHERE ir."Checking" = 'NG'
      AND ih."InspectionHeaderStatus" IN ('Completed', 'Approved')
      AND ih."InspectorID" IS NOT NULL
      AND ih."InspectorID" <> ''
      AND NOT EXISTS (
          SELECT 1
          FROM "Issue" issue_by_result
          WHERE issue_by_result."ResultID" = ir."ResultID"
      )
      AND NOT EXISTS (
          SELECT 1
          FROM "Issue" active_issue
          WHERE active_issue."DetailKawasanID" = ih."DetailKawasanID"
            AND active_issue."IssueStatus" NOT IN ('Closed', 'Verified')
      )
    ORDER BY
        ih."DetailKawasanID",
        ih."InspectionHeaderUpdatedAt" DESC,
        ir."InspectionResultUpdatedAt" DESC
)
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
    'ISS-BF-' || LEFT(MD5("ResultID"), 12),
    "ResultID",
    "DetailKawasanID",
    "InspectorID",
    CURRENT_DATE + 14,
    'Open',
    "Keterangan",
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM missing_ng
ON CONFLICT ("IssueID") DO NOTHING;

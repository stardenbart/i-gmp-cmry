-- Optimize dashboard time-series aggregation without indexing proof photos.
CREATE INDEX IF NOT EXISTS "idx_issue_photo_initial_created"
    ON "Issue_Photo" ("PhotoCreatedAt", "IssueID")
    WHERE "PhotoType" = 'Initial';

CREATE INDEX IF NOT EXISTS "idx_inspection_header_created_area"
    ON "Inspection_Header" ("InspectionHeaderCreatedAt", "AreaID");

CREATE INDEX IF NOT EXISTS "idx_inspection_header_inspector_created"
    ON "Inspection_Header" ("InspectorID", "InspectionHeaderCreatedAt");

CREATE INDEX IF NOT EXISTS "idx_inspection_result_inspection_checking"
    ON "Inspection_Result" ("InspectionID", "Checking");

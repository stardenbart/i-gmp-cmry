-- =========================================================
-- Migration 005: Issue & Follow Up Tables
-- =========================================================

CREATE TABLE IF NOT EXISTS "Issue" (
    "IssueID"        VARCHAR(20) NOT NULL,
    "ResultID"       VARCHAR(20) NOT NULL,
    "IssuePICUserID" VARCHAR(20) NOT NULL,
    "DueDate"        DATE,
    "IssueStatus"    VARCHAR(30) NOT NULL DEFAULT 'Open',
    "Keterangan"     VARCHAR(255),
    "IssueCreatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "IssueUpdatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("IssueID"),
    CONSTRAINT "fk_issue_result"  FOREIGN KEY ("ResultID")       REFERENCES "Inspection_Result" ("ResultID"),
    CONSTRAINT "fk_issue_pic"     FOREIGN KEY ("IssuePICUserID") REFERENCES "Users" ("UserID")
);

CREATE TABLE IF NOT EXISTS "Issue_Photo" (
    "IssuePhotoID"   VARCHAR(20) NOT NULL,
    "IssueID"        VARCHAR(20) NOT NULL,
    "PICUserID"      VARCHAR(20) NOT NULL,
    "PhotoType"      VARCHAR(20) NOT NULL, -- Initial | FollowUp
    "ImageUrl"       VARCHAR(255),
    "FileName"       VARCHAR(255),
    "FollowUpDate"   DATE,
    "JumlahFollowUp" INT,
    "PhotoCreatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "PhotoUpdatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("IssuePhotoID"),
    CONSTRAINT "fk_photo_issue" FOREIGN KEY ("IssueID")   REFERENCES "Issue" ("IssueID"),
    CONSTRAINT "fk_photo_pic"   FOREIGN KEY ("PICUserID") REFERENCES "Users" ("UserID")
);

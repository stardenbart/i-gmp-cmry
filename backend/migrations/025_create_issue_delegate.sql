-- =========================================================
-- Migration 025: Create Issue_Delegate Table
-- =========================================================

CREATE TABLE IF NOT EXISTS "Issue_Delegate" (
    "IssueID"        VARCHAR(30) NOT NULL,
    "DelegateUserID" VARCHAR(20) NOT NULL,
    "DelegatedAt"    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "DelegatedBy"    VARCHAR(20) NOT NULL,
    PRIMARY KEY ("IssueID", "DelegateUserID"),
    CONSTRAINT "fk_issue_delegate_issue" FOREIGN KEY ("IssueID") REFERENCES "Issue" ("IssueID") ON DELETE CASCADE,
    CONSTRAINT "fk_issue_delegate_user"  FOREIGN KEY ("DelegateUserID") REFERENCES "Users" ("UserID") ON DELETE CASCADE,
    CONSTRAINT "fk_issue_delegate_by"    FOREIGN KEY ("DelegatedBy") REFERENCES "Users" ("UserID") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_issue_delegate_user" ON "Issue_Delegate" ("DelegateUserID");

-- =========================================================
-- Migration 012: Create User_Permission Table for User Permission Overrides
-- =========================================================

CREATE TABLE IF NOT EXISTS "User_Permission" (
    "UserPermissionID"        VARCHAR(50) NOT NULL,
    "UserID"                  VARCHAR(50) NOT NULL,
    "PermissionID"            VARCHAR(50) NOT NULL,
    "IsAllowed"               BOOLEAN     NOT NULL DEFAULT TRUE,
    "UserPermissionCreatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UserPermissionUpdatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UserPermissionUpdatedBy" VARCHAR(20),
    PRIMARY KEY ("UserPermissionID"),
    CONSTRAINT "uq_user_permission" UNIQUE ("UserID", "PermissionID"),
    CONSTRAINT "fk_userperm_user"       FOREIGN KEY ("UserID")       REFERENCES "Users" ("UserID") ON DELETE CASCADE,
    CONSTRAINT "fk_userperm_permission" FOREIGN KEY ("PermissionID") REFERENCES "Permission_Master" ("PermissionID") ON DELETE CASCADE
);

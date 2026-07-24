-- =========================================================
-- Migration 006: Logging Tables
-- =========================================================

CREATE TABLE IF NOT EXISTS "Login_Log" (
    "LoginLogID"   VARCHAR(20)  NOT NULL,
    "UserID"       VARCHAR(20),
    "LoginAt"      TIMESTAMP    NOT NULL,
    "LogoutAt"     TIMESTAMP,
    "IPAddress"    VARCHAR(50),
    "DeviceInfo"   VARCHAR(255),
    "LoginStatus"  VARCHAR(20)  NOT NULL, -- Success | Failed
    PRIMARY KEY ("LoginLogID"),
    CONSTRAINT "fk_loginlog_user" FOREIGN KEY ("UserID") REFERENCES "Users" ("UserID")
);

CREATE TABLE IF NOT EXISTS "Activity_Log" (
    "ActivityLogID"       VARCHAR(20)  NOT NULL,
    "UserID"              VARCHAR(20)  NOT NULL,
    "ModuleID"            VARCHAR(20),
    "PermissionID"        VARCHAR(20),
    "ActivityAction"      VARCHAR(50)  NOT NULL,
    "TableAffected"       VARCHAR(100),
    "RecordID"            VARCHAR(50),
    "OldValue"            TEXT,
    "NewValue"            TEXT,
    "ActivityDescription" VARCHAR(255),
    "IPAddress"           VARCHAR(50),
    "ActivityCreatedAt"   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("ActivityLogID"),
    CONSTRAINT "fk_actlog_user"       FOREIGN KEY ("UserID")       REFERENCES "Users" ("UserID"),
    CONSTRAINT "fk_actlog_module"     FOREIGN KEY ("ModuleID")     REFERENCES "Module_Master" ("ModuleID"),
    CONSTRAINT "fk_actlog_permission" FOREIGN KEY ("PermissionID") REFERENCES "Permission_Master" ("PermissionID")
);

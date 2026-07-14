-- =========================================================
-- Migration 002: Auth, Role & Permission Tables
-- =========================================================

CREATE TABLE IF NOT EXISTS "Role_Master" (
    "RoleID"          VARCHAR(20)  NOT NULL,
    "RoleName"        VARCHAR(50)  NOT NULL,
    "RoleDescription" VARCHAR(255),
    "RoleCreatedAt"   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "RoleUpdatedAt"   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("RoleID")
);

CREATE TABLE IF NOT EXISTS "Users" (
    "UserID"           VARCHAR(20)  NOT NULL,
    "DepartmentID"     VARCHAR(20)  NOT NULL,
    "RoleID"           VARCHAR(20)  NOT NULL,
    "Username"         VARCHAR(50)  NOT NULL,
    "FullName"         VARCHAR(100) NOT NULL,
    "Email"            VARCHAR(100) NOT NULL,
    "PasswordHash"     VARCHAR(255) NOT NULL,
    "UserStatus"       VARCHAR(20)  NOT NULL DEFAULT 'Active',
    "UserCreatedAt"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UserUpdatedAt"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("UserID"),
    CONSTRAINT "uq_users_username" UNIQUE ("Username"),
    CONSTRAINT "uq_users_email" UNIQUE ("Email"),
    CONSTRAINT "fk_users_department" FOREIGN KEY ("DepartmentID") REFERENCES "Department_Master" ("DepartmentID"),
    CONSTRAINT "fk_users_role"       FOREIGN KEY ("RoleID")       REFERENCES "Role_Master" ("RoleID")
);

CREATE TABLE IF NOT EXISTS "Module_Master" (
    "ModuleID"        VARCHAR(20)  NOT NULL,
    "ModuleName"      VARCHAR(100) NOT NULL,
    "ModuleCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "ModuleUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("ModuleID")
);

CREATE TABLE IF NOT EXISTS "Permission_Master" (
    "PermissionID"        VARCHAR(20) NOT NULL,
    "ModuleID"            VARCHAR(20) NOT NULL,
    "PermissionCode"      VARCHAR(50) NOT NULL,
    "PermissionName"      VARCHAR(100) NOT NULL,
    "PermissionCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "PermissionUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("PermissionID"),
    CONSTRAINT "fk_permission_module" FOREIGN KEY ("ModuleID") REFERENCES "Module_Master" ("ModuleID")
);

CREATE TABLE IF NOT EXISTS "Role_Permission" (
    "RolePermissionID"        VARCHAR(20) NOT NULL,
    "RoleID"                  VARCHAR(20) NOT NULL,
    "PermissionID"            VARCHAR(20) NOT NULL,
    "IsAllowed"               BOOLEAN     NOT NULL DEFAULT TRUE,
    "RolePermissionCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "RolePermissionUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "RolePermissionUpdatedBy" VARCHAR(20),
    PRIMARY KEY ("RolePermissionID"),
    CONSTRAINT "uq_role_permission" UNIQUE ("RoleID", "PermissionID"),
    CONSTRAINT "fk_roleperm_role"       FOREIGN KEY ("RoleID")       REFERENCES "Role_Master" ("RoleID"),
    CONSTRAINT "fk_roleperm_permission" FOREIGN KEY ("PermissionID") REFERENCES "Permission_Master" ("PermissionID"),
    CONSTRAINT "fk_roleperm_updatedby"  FOREIGN KEY ("RolePermissionUpdatedBy") REFERENCES "Users" ("UserID")
);

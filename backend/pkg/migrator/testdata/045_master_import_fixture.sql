CREATE TABLE "Plant_Master" (
    "PlantID" VARCHAR(50) PRIMARY KEY,
    "PlantName" VARCHAR(100) NOT NULL
);

CREATE TABLE "Users" (
    "UserID" VARCHAR(50) PRIMARY KEY
);

CREATE TABLE "Area_Master" (
    "AreaID" VARCHAR(50) PRIMARY KEY,
    "PlantID" VARCHAR(50) REFERENCES "Plant_Master" ("PlantID"),
    "AreaName" VARCHAR(100) NOT NULL
);

CREATE TABLE "Aspek_Master" (
    "AspekID" VARCHAR(50) PRIMARY KEY,
    "AreaID" VARCHAR(50) NOT NULL REFERENCES "Area_Master" ("AreaID"),
    "AspekName" VARCHAR(100) NOT NULL
);

CREATE TABLE "Detail_Master" (
    "DetailID" VARCHAR(50) PRIMARY KEY,
    "AspekID" VARCHAR(50) NOT NULL REFERENCES "Aspek_Master" ("AspekID"),
    "DetailName" VARCHAR(150) NOT NULL
);

CREATE TABLE "Uraian_Master" (
    "UraianID" VARCHAR(50) PRIMARY KEY,
    "DetailID" VARCHAR(50) NOT NULL REFERENCES "Detail_Master" ("DetailID"),
    "UraianText" VARCHAR(500) NOT NULL,
    "StandardScore" INT NOT NULL DEFAULT 2
);

CREATE TABLE "HEI_Master" (
    "HEIID" VARCHAR(50) PRIMARY KEY,
    "CategoryName" VARCHAR(50) NOT NULL,
    "HEICode" VARCHAR(50),
    "HEIName" VARCHAR(150) NOT NULL,
    "Status" VARCHAR(20) NOT NULL DEFAULT 'Active'
);

CREATE TABLE "Module_Master" (
    "ModuleID" VARCHAR(50) PRIMARY KEY,
    "ModuleName" VARCHAR(100) NOT NULL
);

CREATE TABLE "Permission_Master" (
    "PermissionID" VARCHAR(50) PRIMARY KEY,
    "ModuleID" VARCHAR(50) NOT NULL REFERENCES "Module_Master" ("ModuleID"),
    "PermissionCode" VARCHAR(50) NOT NULL,
    "PermissionName" VARCHAR(100) NOT NULL
);

CREATE TABLE "Role_Master" (
    "RoleID" VARCHAR(50) PRIMARY KEY,
    "RoleName" VARCHAR(50) NOT NULL
);

CREATE TABLE "Role_Permission" (
    "RolePermissionID" VARCHAR(50) PRIMARY KEY,
    "RoleID" VARCHAR(50) NOT NULL REFERENCES "Role_Master" ("RoleID"),
    "PermissionID" VARCHAR(50) NOT NULL REFERENCES "Permission_Master" ("PermissionID"),
    "IsAllowed" BOOLEAN NOT NULL DEFAULT TRUE,
    "RolePermissionUpdatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE ("RoleID", "PermissionID")
);

INSERT INTO "Module_Master" ("ModuleID", "ModuleName") VALUES ('MOD-MSTR', 'Master Data');
INSERT INTO "Role_Master" ("RoleID", "RoleName") VALUES ('ROLE-001', 'Admin');

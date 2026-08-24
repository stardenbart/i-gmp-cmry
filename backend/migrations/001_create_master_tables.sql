-- =========================================================
-- Migration 001: Master Organisasi Tables
-- =========================================================

CREATE TABLE IF NOT EXISTS "Department_Master" (
    "DepartmentID"        VARCHAR(20)  NOT NULL,
    "DepartmentName"      VARCHAR(100) NOT NULL,
    "DepartmentCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "DepartmentUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("DepartmentID")
);

CREATE TABLE IF NOT EXISTS "Area_Master" (
    "AreaID"        VARCHAR(20)  NOT NULL,
    "AreaName"      VARCHAR(100) NOT NULL,
    "AreaCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "AreaUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("AreaID")
);

CREATE TABLE IF NOT EXISTS "Kawasan_Master" (
    "KawasanID"        VARCHAR(20)  NOT NULL,
    "AreaID"           VARCHAR(20)  NOT NULL,
    "KawasanName"      VARCHAR(100) NOT NULL,
    "LastInspection"   TIMESTAMP,
    "KawasanCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "KawasanUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("KawasanID"),
    CONSTRAINT "fk_kawasan_area" FOREIGN KEY ("AreaID") REFERENCES "Area_Master" ("AreaID")
);

CREATE TABLE IF NOT EXISTS "DetailKawasan_Master" (
    "DetailKawasanID"        VARCHAR(20)  NOT NULL,
    "KawasanID"              VARCHAR(20)  NOT NULL,
    "DetailKawasanName"      VARCHAR(100) NOT NULL,
    "LastInspection"         TIMESTAMP,
    "DetailKawasanCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "DetailKawasanUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("DetailKawasanID"),
    CONSTRAINT "fk_detailkawasan_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID")
);

-- =========================================================
-- Migration 004: Inspection Tables
-- =========================================================

CREATE TABLE IF NOT EXISTS "Aspek_Master" (
    "AspekID"        VARCHAR(20)  NOT NULL,
    "AreaID"         VARCHAR(20)  NOT NULL,
    "AspekName"      VARCHAR(100) NOT NULL,
    "AspekCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "AspekUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("AspekID"),
    CONSTRAINT "fk_aspek_area" FOREIGN KEY ("AreaID") REFERENCES "Area_Master" ("AreaID")
);

CREATE TABLE IF NOT EXISTS "Detail_Master" (
    "DetailID"        VARCHAR(20)  NOT NULL,
    "AspekID"         VARCHAR(20)  NOT NULL,
    "DetailName"      VARCHAR(150) NOT NULL,
    "DetailCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "DetailUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("DetailID"),
    CONSTRAINT "fk_detail_aspek" FOREIGN KEY ("AspekID") REFERENCES "Aspek_Master" ("AspekID")
);

CREATE TABLE IF NOT EXISTS "Uraian_Master" (
    "UraianID"        VARCHAR(20)  NOT NULL,
    "DetailID"        VARCHAR(20)  NOT NULL,
    "UraianText"      VARCHAR(500) NOT NULL,
    "StandardScore"   INT          NOT NULL DEFAULT 0,
    "UraianCreatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UraianUpdatedAt" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("UraianID"),
    CONSTRAINT "fk_uraian_detail" FOREIGN KEY ("DetailID") REFERENCES "Detail_Master" ("DetailID")
);

CREATE TABLE IF NOT EXISTS "Inspection_Header" (
    "InspectionID"              VARCHAR(20) NOT NULL,
    "AreaID"                    VARCHAR(20) NOT NULL,
    "KawasanID"                 VARCHAR(20) NOT NULL,
    "DetailKawasanID"           VARCHAR(20) NOT NULL,
    "InspectorID"               VARCHAR(20) NOT NULL,
    "InspectionHeaderStatus"    VARCHAR(30) NOT NULL DEFAULT 'Draft',
    "InspectionHeaderCreatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "InspectionheaderUpdatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("InspectionID"),
    CONSTRAINT "fk_insp_area"    FOREIGN KEY ("AreaID")          REFERENCES "Area_Master" ("AreaID"),
    CONSTRAINT "fk_insp_kawasan" FOREIGN KEY ("KawasanID")       REFERENCES "Kawasan_Master" ("KawasanID"),
    CONSTRAINT "fk_insp_dk"      FOREIGN KEY ("DetailKawasanID") REFERENCES "DetailKawasan_Master" ("DetailKawasanID"),
    CONSTRAINT "fk_insp_inspector" FOREIGN KEY ("InspectorID")   REFERENCES "Users" ("UserID")
);

CREATE TABLE IF NOT EXISTS "Inspection_Result" (
    "ResultID"                   VARCHAR(20)  NOT NULL,
    "InspectionID"               VARCHAR(20)  NOT NULL,
    "UraianID"                   VARCHAR(20)  NOT NULL,
    "Checking"                   VARCHAR(20),
    "Nilai"                      INT          NOT NULL DEFAULT 0,
    "Keterangan"                 VARCHAR(255),
    "InspectionResultCreatedAt"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "InspectionResultUpdatedAt"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("ResultID"),
    CONSTRAINT "fk_result_inspection" FOREIGN KEY ("InspectionID") REFERENCES "Inspection_Header" ("InspectionID"),
    CONSTRAINT "fk_result_uraian"     FOREIGN KEY ("UraianID")     REFERENCES "Uraian_Master" ("UraianID")
);

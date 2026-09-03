-- =========================================================
-- Migration 045: Strict master-data imports and normalized uniqueness
-- =========================================================

-- PostgreSQL is the single source of truth for duplicate identity.  Import
-- preview, commit, regular CRUD, and the unique indexes all call these
-- functions so whitespace/case/unicode rules cannot drift between layers.
CREATE OR REPLACE FUNCTION master_clean_text(input_text TEXT)
RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
PARALLEL SAFE
RETURNS NULL ON NULL INPUT
AS $$
    SELECT regexp_replace(
        btrim(normalize(input_text, NFC)),
        '[[:space:]]+',
        ' ',
        'g'
    );
$$;

CREATE OR REPLACE FUNCTION master_normalize_text(input_text TEXT)
RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
PARALLEL SAFE
RETURNS NULL ON NULL INPUT
AS $$
    SELECT lower(master_clean_text(input_text));
$$;

CREATE OR REPLACE FUNCTION master_normalize_code(input_text TEXT)
RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
PARALLEL SAFE
RETURNS NULL ON NULL INPUT
AS $$
    SELECT upper(master_normalize_text(input_text));
$$;

-- Stop here rather than guessing how existing duplicate records should be
-- merged. Those records may already be referenced by inspections/issues.
DO $$
DECLARE
    duplicate_groups BIGINT;
BEGIN
    SELECT count(*) INTO duplicate_groups
    FROM (
        SELECT "AreaID", master_normalize_text("AspekName")
        FROM "Aspek_Master"
        GROUP BY "AreaID", master_normalize_text("AspekName")
        HAVING count(*) > 1
    ) duplicate_aspek;
    IF duplicate_groups > 0 THEN
        RAISE EXCEPTION 'master import migration blocked: % duplicate Aspek group(s) require manual reconciliation', duplicate_groups;
    END IF;

    SELECT count(*) INTO duplicate_groups
    FROM (
        SELECT "AspekID", master_normalize_text("DetailName")
        FROM "Detail_Master"
        GROUP BY "AspekID", master_normalize_text("DetailName")
        HAVING count(*) > 1
    ) duplicate_detail;
    IF duplicate_groups > 0 THEN
        RAISE EXCEPTION 'master import migration blocked: % duplicate Detail group(s) require manual reconciliation', duplicate_groups;
    END IF;

    SELECT count(*) INTO duplicate_groups
    FROM (
        SELECT "DetailID", master_normalize_text("UraianText")
        FROM "Uraian_Master"
        GROUP BY "DetailID", master_normalize_text("UraianText")
        HAVING count(*) > 1
    ) duplicate_uraian;
    IF duplicate_groups > 0 THEN
        RAISE EXCEPTION 'master import migration blocked: % duplicate Uraian group(s) require manual reconciliation', duplicate_groups;
    END IF;

    SELECT count(*) INTO duplicate_groups
    FROM (
        SELECT master_normalize_text("CategoryName"), master_normalize_text("HEIName")
        FROM "HEI_Master"
        GROUP BY master_normalize_text("CategoryName"), master_normalize_text("HEIName")
        HAVING count(*) > 1
    ) duplicate_hei_name;
    IF duplicate_groups > 0 THEN
        RAISE EXCEPTION 'master import migration blocked: % duplicate HEI name group(s) require manual reconciliation', duplicate_groups;
    END IF;

    SELECT count(*) INTO duplicate_groups
    FROM (
        SELECT master_normalize_text("CategoryName"), master_normalize_code("HEICode")
        FROM "HEI_Master"
        WHERE coalesce(master_normalize_code("HEICode"), '') <> ''
        GROUP BY master_normalize_text("CategoryName"), master_normalize_code("HEICode")
        HAVING count(*) > 1
    ) duplicate_hei_code;
    IF duplicate_groups > 0 THEN
        RAISE EXCEPTION 'master import migration blocked: % duplicate HEI code group(s) require manual reconciliation', duplicate_groups;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS "uq_aspek_area_normalized_name"
    ON "Aspek_Master" ("AreaID", master_normalize_text("AspekName"));

CREATE UNIQUE INDEX IF NOT EXISTS "uq_detail_aspek_normalized_name"
    ON "Detail_Master" ("AspekID", master_normalize_text("DetailName"));

CREATE UNIQUE INDEX IF NOT EXISTS "uq_uraian_detail_normalized_text"
    ON "Uraian_Master" ("DetailID", master_normalize_text("UraianText"));

CREATE UNIQUE INDEX IF NOT EXISTS "uq_hei_category_normalized_name"
    ON "HEI_Master" (master_normalize_text("CategoryName"), master_normalize_text("HEIName"));

CREATE UNIQUE INDEX IF NOT EXISTS "uq_hei_category_normalized_code"
    ON "HEI_Master" (master_normalize_text("CategoryName"), master_normalize_code("HEICode"))
    WHERE coalesce(master_normalize_code("HEICode"), '') <> '';

CREATE TABLE IF NOT EXISTS "Master_Import_Batch" (
    "ImportID"             VARCHAR(50)  NOT NULL,
    "UserID"               VARCHAR(50)  NOT NULL,
    "PlantID"              VARCHAR(50),
    "ImportType"           VARCHAR(20)  NOT NULL,
    "OriginalFileName"     VARCHAR(255) NOT NULL,
    "FileHash"             VARCHAR(64)  NOT NULL,
    "TemplateVersion"      VARCHAR(20)  NOT NULL,
    "TotalRows"            INT          NOT NULL DEFAULT 0,
    "ValidRows"            INT          NOT NULL DEFAULT 0,
    "InvalidRows"          INT          NOT NULL DEFAULT 0,
    "InsertedRows"         INT          NOT NULL DEFAULT 0,
    "Status"               VARCHAR(30)  NOT NULL,
    "ErrorSummary"         TEXT,
    "CreatedAt"            TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "ValidatedAt"          TIMESTAMP,
    "CommittedAt"          TIMESTAMP,
    "UpdatedAt"            TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("ImportID"),
    CONSTRAINT "fk_master_import_user" FOREIGN KEY ("UserID") REFERENCES "Users" ("UserID"),
    CONSTRAINT "fk_master_import_plant" FOREIGN KEY ("PlantID") REFERENCES "Plant_Master" ("PlantID") ON DELETE SET NULL,
    CONSTRAINT "chk_master_import_type" CHECK ("ImportType" IN ('aspek', 'detail', 'uraian', 'hei')),
    CONSTRAINT "chk_master_import_status" CHECK ("Status" IN ('Uploaded', 'ValidationRejected', 'Validated', 'Committing', 'Committed', 'Failed', 'Expired'))
);

CREATE INDEX IF NOT EXISTS "idx_master_import_actor_created"
    ON "Master_Import_Batch" ("UserID", "CreatedAt" DESC);
CREATE INDEX IF NOT EXISTS "idx_master_import_hash"
    ON "Master_Import_Batch" ("FileHash", "ImportType", "PlantID");
CREATE INDEX IF NOT EXISTS "idx_master_import_status_updated"
    ON "Master_Import_Batch" ("Status", "UpdatedAt");

CREATE TABLE IF NOT EXISTS "Master_Import_Row" (
    "ImportID"  VARCHAR(50) NOT NULL,
    "EntityType" VARCHAR(20) NOT NULL,
    "EntityID"  VARCHAR(50) NOT NULL,
    "SourceRow" INT         NOT NULL,
    PRIMARY KEY ("ImportID", "EntityType", "EntityID"),
    CONSTRAINT "fk_master_import_row_batch" FOREIGN KEY ("ImportID") REFERENCES "Master_Import_Batch" ("ImportID") ON DELETE CASCADE
);

INSERT INTO "Permission_Master" ("PermissionID", "ModuleID", "PermissionCode", "PermissionName")
VALUES ('PERM-MSTR-I', 'MOD-MSTR', 'IMPORT', 'Import Master Data')
ON CONFLICT ("PermissionID") DO UPDATE SET
    "ModuleID" = EXCLUDED."ModuleID",
    "PermissionCode" = EXCLUDED."PermissionCode",
    "PermissionName" = EXCLUDED."PermissionName";

INSERT INTO "Role_Permission" ("RolePermissionID", "RoleID", "PermissionID", "IsAllowed")
VALUES ('RP-ADM-MSTR-I', 'ROLE-001', 'PERM-MSTR-I', TRUE)
ON CONFLICT ("RoleID", "PermissionID") DO UPDATE SET
    "IsAllowed" = TRUE,
    "RolePermissionUpdatedAt" = CURRENT_TIMESTAMP;

-- =========================================================
-- Migration 003: PIC Mapping Table
-- =========================================================

CREATE TABLE IF NOT EXISTS "PIC_Mapping" (
    "PICMapID"            VARCHAR(50) NOT NULL,
    "AreaID"              VARCHAR(50),
    "KawasanID"           VARCHAR(50) NOT NULL,
    "UserID"              VARCHAR(50) NOT NULL,
    "KategoriPIC"         VARCHAR(50),
    "PICMappingCreatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "PICMappingUpdatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("PICMapID"),
    CONSTRAINT "fk_picmap_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID"),
    CONSTRAINT "fk_picmap_user"    FOREIGN KEY ("UserID")    REFERENCES "Users" ("UserID")
);

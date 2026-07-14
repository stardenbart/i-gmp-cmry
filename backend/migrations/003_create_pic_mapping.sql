-- =========================================================
-- Migration 003: PIC Mapping Table
-- =========================================================

CREATE TABLE IF NOT EXISTS "PIC_Mapping" (
    "PICMapID"            VARCHAR(20) NOT NULL,
    "AreaID"              VARCHAR(20) NOT NULL,
    "KawasanID"           VARCHAR(20) NOT NULL,
    "UserID"              VARCHAR(20) NOT NULL,
    "KategoriPIC"         VARCHAR(50),
    "PICMappingCreatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "PICMappingUpdatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("PICMapID"),
    CONSTRAINT "fk_picmap_area"    FOREIGN KEY ("AreaID")    REFERENCES "Area_Master" ("AreaID"),
    CONSTRAINT "fk_picmap_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID"),
    CONSTRAINT "fk_picmap_user"    FOREIGN KEY ("UserID")    REFERENCES "Users" ("UserID")
);

-- =========================================================
-- Migration 017: Distributed Inspection System Tables
-- =========================================================

CREATE TABLE IF NOT EXISTS "kawasan_aspek" (
    "KawasanAspekID" VARCHAR(36) NOT NULL,
    "KawasanID"       VARCHAR(20) NOT NULL,
    "AspekID"         VARCHAR(20) NOT NULL,
    "IsRequired"      BOOLEAN     NOT NULL DEFAULT TRUE,
    "CreatedAt"       TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("KawasanAspekID"),
    CONSTRAINT "fk_ka_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID") ON DELETE CASCADE,
    CONSTRAINT "fk_ka_aspek"   FOREIGN KEY ("AspekID")   REFERENCES "Aspek_Master" ("AspekID") ON DELETE CASCADE,
    CONSTRAINT "unique_kawasan_aspek" UNIQUE ("KawasanID", "AspekID")
);

CREATE TABLE IF NOT EXISTS "inspeksi_session" (
    "SessionID"      VARCHAR(36) NOT NULL,
    "KawasanID"      VARCHAR(20) NOT NULL,
    "Periode"        DATE        NOT NULL,
    "Status"         VARCHAR(30) NOT NULL DEFAULT 'InProgress',
    "TotalAspek"     INT         NOT NULL DEFAULT 0,
    "CompletedAspek" INT         NOT NULL DEFAULT 0,
    "SyncedAt"       TIMESTAMP,
    "CreatedAt"      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt"      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("SessionID"),
    CONSTRAINT "fk_session_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID"),
    CONSTRAINT "unique_session_kawasan_periode" UNIQUE ("KawasanID", "Periode")
);

CREATE TABLE IF NOT EXISTS "inspeksi_aspek_result" (
    "ResultID"     VARCHAR(36) NOT NULL,
    "SessionID"    VARCHAR(36) NOT NULL,
    "KawasanID"    VARCHAR(20) NOT NULL,
    "AspekID"      VARCHAR(20) NOT NULL,
    "UserID"       VARCHAR(20) NOT NULL,
    "DataInspeksi" JSONB       NOT NULL,
    "Skor"         NUMERIC(5,2),
    "Catatan"      TEXT,
    "CompletedAt"  TIMESTAMP   NOT NULL,
    "CreatedAt"    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("ResultID"),
    CONSTRAINT "fk_iar_session" FOREIGN KEY ("SessionID") REFERENCES "inspeksi_session" ("SessionID") ON DELETE CASCADE,
    CONSTRAINT "fk_iar_kawasan" FOREIGN KEY ("KawasanID") REFERENCES "Kawasan_Master" ("KawasanID"),
    CONSTRAINT "fk_iar_aspek"   FOREIGN KEY ("AspekID")   REFERENCES "Aspek_Master" ("AspekID"),
    CONSTRAINT "fk_iar_user"    FOREIGN KEY ("UserID")    REFERENCES "Users" ("UserID"),
    CONSTRAINT "unique_session_aspek" UNIQUE ("SessionID", "AspekID")
);

CREATE TABLE IF NOT EXISTS "inspeksi_audit_log" (
    "LogID"     VARCHAR(36) NOT NULL,
    "SessionID" VARCHAR(36) NOT NULL,
    "AspekID"   VARCHAR(20) NOT NULL,
    "UserID"    VARCHAR(20) NOT NULL,
    "Action"    VARCHAR(50) NOT NULL,
    "Meta"      JSONB,
    "CreatedAt" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("LogID")
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS "idx_session_kawasan_periode" ON "inspeksi_session" ("KawasanID", "Periode");
CREATE INDEX IF NOT EXISTS "idx_result_session" ON "inspeksi_aspek_result" ("SessionID");
CREATE INDEX IF NOT EXISTS "idx_result_kawasan_aspek" ON "inspeksi_aspek_result" ("KawasanID", "AspekID");
CREATE INDEX IF NOT EXISTS "idx_session_status" ON "inspeksi_session" ("Status") WHERE "Status" != 'Synced';
CREATE INDEX IF NOT EXISTS "idx_audit_session" ON "inspeksi_audit_log" ("SessionID", "CreatedAt" DESC);

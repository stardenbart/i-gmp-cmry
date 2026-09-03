CREATE TABLE IF NOT EXISTS "KPI_Public_Share" (
    "ShareID" VARCHAR(50) PRIMARY KEY,
    "OwnerUserID" VARCHAR(50) NOT NULL REFERENCES "Users"("UserID") ON DELETE CASCADE,
    "ShareName" VARCHAR(100) NOT NULL,
    "PublicTitle" VARCHAR(150) NOT NULL,
    "TokenHash" CHAR(64) NOT NULL UNIQUE,
    "TokenPrefix" VARCHAR(16) NOT NULL,
    "PlantID" VARCHAR(50) NOT NULL REFERENCES "Plant_Master"("PlantID") ON DELETE RESTRICT,
    "LayoutJSON" JSONB NOT NULL,
    "FilterJSON" JSONB NOT NULL DEFAULT '{}'::jsonb,
    "AllowPeriodChange" BOOLEAN NOT NULL DEFAULT FALSE,
    "SnapshotVersion" SMALLINT NOT NULL DEFAULT 1,
    "ExpiresAt" TIMESTAMPTZ,
    "RevokedAt" TIMESTAMPTZ,
    "LastAccessedAt" TIMESTAMPTZ,
    "AccessCount" BIGINT NOT NULL DEFAULT 0,
    "CreatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "UpdatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_kpi_public_share_owner" ON "KPI_Public_Share" ("OwnerUserID", "CreatedAt" DESC);
CREATE INDEX IF NOT EXISTS "idx_kpi_public_share_plant" ON "KPI_Public_Share" ("PlantID");
CREATE INDEX IF NOT EXISTS "idx_kpi_public_share_expiry" ON "KPI_Public_Share" ("ExpiresAt") WHERE "RevokedAt" IS NULL;


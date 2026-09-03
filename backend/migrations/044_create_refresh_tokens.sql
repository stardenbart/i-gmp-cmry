-- Refresh_Token: opaque, revocable session tokens backing the httpOnly
-- cookie-based auth flow (see backend/pkg/refreshtoken, authusecase.Refresh).
-- Only a SHA-256 hash of each token is stored; the raw value never touches
-- the database.
CREATE TABLE IF NOT EXISTS "Refresh_Token" (
    "ID"          VARCHAR(30)  PRIMARY KEY,
    "UserID"      VARCHAR(30)  NOT NULL REFERENCES "Users"("UserID") ON DELETE CASCADE,
    "FamilyID"    VARCHAR(30)  NOT NULL,
    "TokenHash"   VARCHAR(64)  NOT NULL UNIQUE,
    "ExpiresAt"   TIMESTAMPTZ  NOT NULL,
    "RevokedAt"   TIMESTAMPTZ,
    "ReplacedBy"  VARCHAR(30),
    "IPAddress"   VARCHAR(50)  NOT NULL DEFAULT '',
    "DeviceInfo"  VARCHAR(255) NOT NULL DEFAULT '',
    "CreatedAt"   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_refresh_token_user_id ON "Refresh_Token" ("UserID");
CREATE INDEX IF NOT EXISTS idx_refresh_token_family_id ON "Refresh_Token" ("FamilyID");
-- Speeds up the periodic cleanup of expired/revoked rows (not yet automated;
-- safe to prune manually with `DELETE ... WHERE "ExpiresAt" < now()`).
CREATE INDEX IF NOT EXISTS idx_refresh_token_expires_at ON "Refresh_Token" ("ExpiresAt");

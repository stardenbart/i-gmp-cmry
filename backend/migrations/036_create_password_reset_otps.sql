-- One-time password challenges for the two-step forgot-password flow.
-- Only bcrypt hashes are stored; the six-digit OTP is delivered by email.
CREATE TABLE IF NOT EXISTS "Password_Reset_OTP" (
    "ResetID"     VARCHAR(32)  PRIMARY KEY,
    "UserID"      VARCHAR(20)  NOT NULL,
    "OTPHash"     VARCHAR(255) NOT NULL,
    "ExpiresAt"   TIMESTAMP    NOT NULL,
    "Attempts"    INTEGER      NOT NULL DEFAULT 0,
    "UsedAt"      TIMESTAMP,
    "CreatedAt"   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "fk_password_reset_otp_user"
        FOREIGN KEY ("UserID") REFERENCES "Users" ("UserID") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_password_reset_otp_active"
    ON "Password_Reset_OTP" ("UserID", "CreatedAt" DESC)
    WHERE "UsedAt" IS NULL;

CREATE INDEX IF NOT EXISTS "idx_password_reset_otp_expiry"
    ON "Password_Reset_OTP" ("ExpiresAt");

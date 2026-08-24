-- Dynamic global SMTP configuration. Empty values intentionally fall back to
-- environment variables until a Super Admin saves the SMTP form.
ALTER TABLE "System_Setting"
    ADD COLUMN IF NOT EXISTS "PlantID" VARCHAR(50) NOT NULL DEFAULT '';

ALTER TABLE "System_Setting" DROP CONSTRAINT IF EXISTS "System_Setting_pkey";
ALTER TABLE "System_Setting"
    ADD CONSTRAINT "System_Setting_pkey" PRIMARY KEY ("SettingKey", "PlantID");

INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
VALUES
    ('SMTP_ENABLED', '', 'true', FALSE, 'Aktifkan atau nonaktifkan pengiriman email SMTP'),
    ('SMTP_HOST', '', '', FALSE, 'Hostname server SMTP; kosong menggunakan environment'),
    ('SMTP_PORT', '', '', FALSE, 'Port server SMTP; kosong menggunakan environment'),
    ('SMTP_USER', '', '', FALSE, 'Username autentikasi SMTP; kosong menggunakan environment'),
    ('SMTP_PASSWORD', '', '', TRUE, 'Password atau app password SMTP (terenkripsi)'),
    ('SMTP_SENDER_EMAIL', '', '', FALSE, 'Alamat email pengirim; kosong menggunakan environment')
ON CONFLICT ("SettingKey", "PlantID") DO NOTHING;

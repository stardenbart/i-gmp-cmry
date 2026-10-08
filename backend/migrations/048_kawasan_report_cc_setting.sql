-- CC list for the Kawasan inspection report email (score per Detail Kawasan
-- + issue list), which replaces the Power Automate GMP_Report_Flow. Global
-- default is empty; Plant Sentul starts with the flow's fixed CC recipients.
-- Admins edit it per plant in Pengaturan > Pengingat & Notifikasi.
INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
VALUES
    (
        'EMAIL_CC_KAWASAN_REPORT',
        '',
        '',
        FALSE,
        'Alamat email yang selalu di-CC pada laporan inspeksi kawasan, dipisahkan titik koma (;)'
    )
ON CONFLICT ("SettingKey", "PlantID") DO NOTHING;

INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
SELECT
    'EMAIL_CC_KAWASAN_REPORT',
    p."PlantID",
    'hera.narulita@cimory.com; risma.dwi@cimory.com; agus.mulyono@cimory.com; nadya.audyra@cimory.com; m.syahril@cimory.com',
    FALSE,
    'Alamat email yang selalu di-CC pada laporan inspeksi kawasan, dipisahkan titik koma (;)'
FROM "Plant_Master" p
WHERE p."PlantID" = 'PLT-SENTUL'
ON CONFLICT ("SettingKey", "PlantID") DO NOTHING;

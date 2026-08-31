-- New setting key for the Kawasan-level "fully inspected" notification.
-- SeedSettings() only runs on fresh installs, so an already-deployed
-- database needs this row inserted explicitly for the feature to have a
-- template to send at all (FindByKey would otherwise return not-found).
INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
VALUES
    (
        'EMAIL_TEMPLATE_KAWASAN_CONFIRMED',
        '',
        '<h1>Kawasan Selesai Diinspeksi</h1><p>Kawasan {{.KawasanID}} telah selesai diinspeksi bulan ini ({{.CompletedDetailKawasan}}/{{.TotalDetailKawasan}} detail kawasan).</p>',
        FALSE,
        'Template email ke Manager saat seluruh Detail Kawasan di satu Kawasan selesai diinspeksi bulan ini (HTML)'
    )
ON CONFLICT ("SettingKey", "PlantID") DO NOTHING;

-- Admin-configurable "inspection period" cutoff day, per plant. Default '1'
-- means an exact calendar month — identical to the previous hardcoded
-- behavior, so no plant's cycle changes until an Admin explicitly sets a
-- different cutoff day for that plant via Settings.
-- SeedSettings() only runs on fresh installs, so an already-deployed
-- database needs this row inserted explicitly.
INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
VALUES
    (
        'INSPECTION_PERIOD_CUTOFF_DAY',
        '',
        '1',
        FALSE,
        'Tanggal berapa siklus bulanan inspeksi dimulai (1-28). Isi 13 supaya periode 13 Jan-12 Feb dihitung sebagai bulan Januari. Isi 1 untuk bulan kalender biasa (default). Bisa diatur berbeda per plant.'
    )
ON CONFLICT ("SettingKey", "PlantID") DO NOTHING;

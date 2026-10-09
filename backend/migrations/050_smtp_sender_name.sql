-- Display name for outgoing email ("Name" <address>), editable per plant in
-- Pengaturan > SMTP. Empty keeps the bare sender address.
INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
VALUES
    ('SMTP_SENDER_NAME', '', '', FALSE,
     'Nama pengirim yang tampil di kotak masuk penerima (mis. I-GMP Notification); kosong hanya menampilkan alamat email')
ON CONFLICT ("SettingKey", "PlantID") DO NOTHING;

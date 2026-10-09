-- On/off switches for template emails (Pengaturan > Template Email). Only
-- the email is switched; in-app notifications are always sent. Defaults to
-- enabled so nothing changes until an admin turns one off. The password
-- reset OTP email has no switch (it is required to reset a password).
INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
VALUES
    ('EMAIL_TEMPLATE_ISSUE_ASSIGNMENT_ENABLED', '', 'true', FALSE,
     'Kirim email penugasan temuan (issue) ke PIC. Notifikasi di aplikasi tetap dikirim.'),
    ('EMAIL_TEMPLATE_INSPECTION_CONFIRMED_ENABLED', '', 'true', FALSE,
     'Kirim email saat inspeksi seluruh area selesai. Notifikasi di aplikasi tetap dikirim.'),
    ('EMAIL_TEMPLATE_DEADLINE_REMINDER_ENABLED', '', 'true', FALSE,
     'Kirim email pengingat tenggat temuan. Notifikasi di aplikasi tetap dikirim.')
ON CONFLICT ("SettingKey", "PlantID") DO NOTHING;

-- Deadline reminder worker needs a per-issue "already reminded" marker so
-- it never re-sends the same reminder every hourly tick.
ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "DeadlineReminderSentAt" TIMESTAMP NULL;

-- New settings for: (1) the deadline-reminder lead time (admin-configurable,
-- "0" disables it), (2) toggles for the Kawasan/Area "fully inspected" email
-- notifications, and (3) a customizable template for the deadline reminder.
-- SeedSettings() only runs on fresh installs, so an already-deployed database
-- needs these rows inserted explicitly.
INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
VALUES
    (
        'ISSUE_DEADLINE_REMINDER_DAYS_BEFORE',
        '',
        '3',
        FALSE,
        'Berapa hari sebelum tenggat temuan sistem mengirim pengingat ke PIC & Manager (0 = nonaktif)'
    ),
    (
        'NOTIFY_ON_KAWASAN_COMPLETE',
        '',
        'true',
        FALSE,
        'Kirim email ke PIC & Manager saat seluruh Detail Kawasan di satu Kawasan selesai diinspeksi bulan ini'
    ),
    (
        'NOTIFY_ON_AREA_COMPLETE',
        '',
        'true',
        FALSE,
        'Kirim email ke PIC & Manager saat seluruh Kawasan di satu Area selesai diinspeksi'
    ),
    (
        'EMAIL_TEMPLATE_DEADLINE_REMINDER',
        '',
        '<h1>Pengingat Tenggat Temuan</h1><p>Halo {{.PICName}}, temuan {{.IssueID}} ({{.Keterangan}}) akan jatuh tempo pada {{.DueDate}} ({{.DaysRemaining}} hari lagi). Mohon segera ditindaklanjuti.</p>',
        FALSE,
        'Template email pengingat sebelum tenggat penyelesaian temuan, ke PIC & Manager (HTML)'
    )
ON CONFLICT ("SettingKey", "PlantID") DO NOTHING;

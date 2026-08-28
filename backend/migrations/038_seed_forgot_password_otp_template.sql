-- Replace only the obsolete seeded forgot-password template. Custom templates
-- remain untouched, while fresh and upgraded installations receive an OTP
-- template that exposes every variable supported by the reset flow.
INSERT INTO "System_Setting"
    ("SettingKey", "PlantID", "SettingValue", "IsEncrypted", "Description")
VALUES
    (
        'EMAIL_TEMPLATE_FORGOT_PASSWORD',
        '',
        '<!DOCTYPE html>
<html lang="id">
<head><meta charset="UTF-8"><title>Reset Password</title></head>
<body style="font-family:Arial,sans-serif;background-color:#f4f4f4;padding:20px;">
  <div style="max-width:600px;margin:auto;background:#ffffff;border-radius:8px;padding:32px;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
    <h2 style="color:#1a1a2e;">Kode OTP Reset Password</h2>
    <p>Halo, <strong>{{.FullName}}</strong>.</p>
    <p>Kami menerima permintaan reset password untuk akun Anda (<code>{{.Username}}</code>).</p>
    <p>Masukkan kode OTP berikut pada halaman reset password:</p>
    <div style="background:#f0f0f0;border-left:4px solid #6c63ff;padding:12px 20px;border-radius:4px;margin:16px 0;">
      <span style="font-size:22px;font-weight:bold;letter-spacing:4px;color:#1a1a2e;">{{.OTP}}</span>
    </div>
    <p>Kode ini berlaku selama <strong>{{.OTPExpiryMinutes}} menit</strong> dan hanya dapat digunakan satu kali.</p>
    <p>Jangan berikan kode OTP ini kepada siapa pun.</p>
    <p>Jika Anda tidak merasa melakukan permintaan ini, abaikan email ini.</p>
    <hr style="border:none;border-top:1px solid #eee;margin:24px 0;">
    <p style="font-size:12px;color:#888;">Email ini dikirim secara otomatis. Jangan balas email ini.</p>
    <p style="font-size:12px;color:#888;">© Monitoring Audit System</p>
  </div>
</body>
</html>',
        FALSE,
        'Template email OTP untuk lupa password (HTML)'
    )
ON CONFLICT ("SettingKey", "PlantID") DO UPDATE
SET
    "SettingValue" = EXCLUDED."SettingValue",
    "Description" = EXCLUDED."Description"
WHERE "System_Setting"."SettingValue" = ''
   OR "System_Setting"."SettingValue" = '<h1>Reset Password</h1><p>Klik tombol berikut untuk mereset password Anda.</p>';

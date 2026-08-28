-- Plant overrides created from the obsolete seeded template hide the new OTP
-- global fallback. Remove only the exact legacy/empty values so customized
-- plant templates remain untouched and can continue overriding the fallback.
DELETE FROM "System_Setting"
WHERE "SettingKey" = 'EMAIL_TEMPLATE_FORGOT_PASSWORD'
  AND "PlantID" <> ''
  AND (
      "SettingValue" = ''
      OR "SettingValue" = '<h1>Reset Password</h1><p>Klik tombol berikut untuk mereset password Anda.</p>'
  );

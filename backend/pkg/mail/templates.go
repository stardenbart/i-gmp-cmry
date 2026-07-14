package mail

// ── Email Templates ─────────────────────────────────────────────────────────

// TmplForgotPassword is the HTML template for "Forgot Password" email.
const TmplForgotPassword = `<!DOCTYPE html>
<html lang="id">
<head><meta charset="UTF-8"><title>Reset Password</title></head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
  <div style="max-width:600px;margin:auto;background:#ffffff;border-radius:8px;padding:32px;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
    <h2 style="color:#1a1a2e;">Reset Password Akun Anda</h2>
    <p>Halo, <strong>{{.FullName}}</strong>.</p>
    <p>Kami menerima permintaan reset password untuk akun Anda (<code>{{.Username}}</code>).</p>
    <p>Berikut adalah <strong>password sementara</strong> Anda:</p>
    <div style="background:#f0f0f0;border-left:4px solid #6c63ff;padding:12px 20px;border-radius:4px;margin:16px 0;">
      <span style="font-size:22px;font-weight:bold;letter-spacing:4px;color:#1a1a2e;">{{.TempPassword}}</span>
    </div>
    <p><strong>Segera ganti password</strong> setelah login menggunakan password di atas melalui menu <em>Change Password</em>.</p>
    <p>Jika Anda tidak merasa melakukan permintaan ini, abaikan email ini.</p>
    <hr style="border:none;border-top:1px solid #eee;margin:24px 0;">
    <p style="font-size:12px;color:#888;">Email ini dikirim secara otomatis. Jangan balas email ini.</p>
    <p style="font-size:12px;color:#888;">© Monitoring Audit System</p>
  </div>
</body>
</html>`

// TmplIssueAssignment is the HTML template for "New Issue Assigned" email.
const TmplIssueAssignment = `<!DOCTYPE html>
<html lang="id">
<head><meta charset="UTF-8"><title>Issue Baru Ditugaskan</title></head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
  <div style="max-width:600px;margin:auto;background:#ffffff;border-radius:8px;padding:32px;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
    <h2 style="color:#1a1a2e;">Issue Baru Ditugaskan kepada Anda</h2>
    <p>Halo, <strong>{{.PICName}}</strong>.</p>
    <p>Sebuah issue baru dari hasil inspeksi telah ditugaskan kepada Anda.</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr style="background:#f8f8ff;">
        <td style="padding:10px 14px;border:1px solid #e0e0e0;font-weight:bold;width:140px;">Issue ID</td>
        <td style="padding:10px 14px;border:1px solid #e0e0e0;"><code>{{.IssueID}}</code></td>
      </tr>
      <tr>
        <td style="padding:10px 14px;border:1px solid #e0e0e0;font-weight:bold;">Keterangan</td>
        <td style="padding:10px 14px;border:1px solid #e0e0e0;">{{.Keterangan}}</td>
      </tr>
      <tr style="background:#f8f8ff;">
        <td style="padding:10px 14px;border:1px solid #e0e0e0;font-weight:bold;">Status</td>
        <td style="padding:10px 14px;border:1px solid #e0e0e0;"><span style="color:#e67e22;font-weight:bold;">{{.Status}}</span></td>
      </tr>
      <tr>
        <td style="padding:10px 14px;border:1px solid #e0e0e0;font-weight:bold;">Due Date</td>
        <td style="padding:10px 14px;border:1px solid #e0e0e0;">{{.DueDate}}</td>
      </tr>
    </table>
    <p>Silakan login ke sistem untuk menangani issue ini sesegera mungkin.</p>
    <hr style="border:none;border-top:1px solid #eee;margin:24px 0;">
    <p style="font-size:12px;color:#888;">Email ini dikirim secara otomatis. Jangan balas email ini.</p>
    <p style="font-size:12px;color:#888;">© Monitoring Audit System</p>
  </div>
</body>
</html>`

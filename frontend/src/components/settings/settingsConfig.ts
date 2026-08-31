export const EMAIL_TEMPLATES = [
  { key: "EMAIL_TEMPLATE_FORGOT_PASSWORD", title: "OTP Lupa Password", description: "Template kode OTP reset password. Dapat dibuat berbeda untuk setiap plant.", variables: ["FullName", "Username", "OTP", "OTPExpiryMinutes"], requiredVariables: ["OTP", "OTPExpiryMinutes"] },
  { key: "EMAIL_TEMPLATE_ISSUE_ASSIGNMENT", title: "Penugasan Temuan (Issue)", description: "Email notifikasi saat user ditugaskan memperbaiki suatu temuan.", variables: ["PICName", "IssueID", "Keterangan", "Status", "DueDate"], requiredVariables: [] },
  { key: "EMAIL_TEMPLATE_INSPECTION_CONFIRMED", title: "Konfirmasi Inspeksi", description: "Email saat inspeksi area telah selesai dan dikonfirmasi.", variables: ["AreaID", "TotalDetailKawasan", "CompletedDetailKawasan", "Status"], requiredVariables: [] },
  { key: "EMAIL_TEMPLATE_KAWASAN_CONFIRMED", title: "Kawasan Selesai Diinspeksi", description: "Email ke Manager saat seluruh Detail Kawasan di satu Kawasan selesai diinspeksi bulan ini.", variables: ["KawasanID", "TotalDetailKawasan", "CompletedDetailKawasan", "Status"], requiredVariables: [] },
] as const;

export type EmailTemplateConfig = (typeof EMAIL_TEMPLATES)[number];

export const GENERAL_SETTINGS = [
  { group: "Keamanan & Akses", keys: [
    { key: "MAX_LOGIN_ATTEMPTS", label: "Maksimal Percobaan Login", type: "number", desc: "Jumlah maksimal percobaan sebelum akun terkunci sementara." },
    { key: "SESSION_IDLE_TIMEOUT_MINUTES", label: "Batas Waktu Sesi (Menit)", type: "number", desc: "Waktu tidak aktif sebelum pengguna dikeluarkan otomatis." },
  ] },
  { group: "Sistem & Penyimpanan", keys: [
    { key: "MAX_UPLOAD_SIZE_MB", label: "Maksimal Ukuran Unggahan (MB)", type: "number", desc: "Batas ukuran maksimal untuk setiap file yang diunggah." },
    { key: "MINIO_ALLOWED_IPS", label: "IP Penyimpanan yang Diizinkan", type: "text", desc: "Daftar IP yang diizinkan mengakses Minio, dipisahkan koma." },
  ] },
  { group: "Tenggat Waktu Temuan", keys: [
    { key: "ISSUE_DEADLINE_DAYS", label: "Tenggat Penyelesaian (Hari)", type: "number", desc: "Waktu penyelesaian default sebelum temuan menjadi overdue." },
    { key: "ISSUE_AUTO_APPROVE_DAYS", label: "Auto-Approve Follow-Up (Hari)", type: "number", desc: "Waktu tunggu peninjauan follow-up sebelum disetujui otomatis." },
    { key: "ISSUE_WOWR_AUTO_APPROVE_DAYS", label: "Auto-Approve WO/WR (Hari)", type: "number", desc: "Waktu tunggu validasi bukti WO/WR sebelum disetujui otomatis." },
  ] },
] as const;

export const SMTP_DEFAULTS: Record<string, string> = {
  SMTP_ENABLED: "true",
  SMTP_HOST: "",
  SMTP_PORT: "",
  SMTP_USER: "",
  SMTP_PASSWORD: "",
  SMTP_SENDER_EMAIL: "",
};

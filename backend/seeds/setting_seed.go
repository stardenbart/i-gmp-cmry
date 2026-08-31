package seeds

import (
	"log"

	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/mail"
	"gorm.io/gorm"
)

// SeedSettings populates initial system settings
func SeedSettings(db *gorm.DB) {
	settings := []master.Setting{
		{
			SettingKey:   master.SettingKeyMinioAllowedIPs,
			SettingValue: "*",
			Description:  "Daftar IP yang diizinkan untuk bypass URL MinIO (pisahkan dengan koma, atau * untuk publik)",
		},
		{
			SettingKey:   master.SettingKeyMaxUploadSizeMB,
			SettingValue: "10",
			Description:  "Batas maksimal ukuran file upload (dalam MB)",
		},
		{
			SettingKey:   master.SettingKeyMaxLoginAttempts,
			SettingValue: "5",
			Description:  "Batas maksimal percobaan login gagal sebelum akun terkunci",
		},
		{
			SettingKey:   master.SettingKeySessionIdleTimeout,
			SettingValue: "30",
			Description:  "Batas waktu sesi idle (dalam menit)",
		},
		{
			SettingKey:   master.SettingKeyEmailTemplateForgotPass,
			SettingValue: mail.TmplForgotPassword,
			Description:  "Template email OTP untuk lupa password (HTML)",
		},
		{
			SettingKey:   master.SettingKeyEmailTemplateIssue,
			SettingValue: "<h1>Tugas Baru</h1><p>Anda mendapatkan tugas baru. Silakan periksa dashboard.</p>",
			Description:  "Template email untuk penugasan issue (HTML)",
		},
		{
			SettingKey:   master.SettingKeyEmailTemplateInspectionConfirmed,
			SettingValue: "<h1>Inspeksi Selesai</h1><p>Semua kawasan pada area ini telah selesai diinspeksi.</p>",
			Description:  "Template email notifikasi inspeksi selesai ke PIC/Manager (HTML)",
		},
		{
			SettingKey:   master.SettingKeyEmailTemplateKawasanConfirmed,
			SettingValue: "<h1>Kawasan Selesai Diinspeksi</h1><p>Kawasan {{.KawasanID}} telah selesai diinspeksi bulan ini ({{.CompletedDetailKawasan}}/{{.TotalDetailKawasan}} detail kawasan).</p>",
			Description:  "Template email ke Manager saat seluruh Detail Kawasan di satu Kawasan selesai diinspeksi bulan ini (HTML)",
		},
		{
			SettingKey:   master.SettingKeyEmailTemplateDeadlineReminder,
			SettingValue: "<h1>Pengingat Tenggat Temuan</h1><p>Halo {{.PICName}}, temuan {{.IssueID}} ({{.Keterangan}}) akan jatuh tempo pada {{.DueDate}} ({{.DaysRemaining}} hari lagi). Mohon segera ditindaklanjuti.</p>",
			Description:  "Template email pengingat sebelum tenggat penyelesaian temuan, ke PIC & Manager (HTML)",
		},
		{
			SettingKey:   master.SettingKeyIssueDeadlineReminderDaysBefore,
			SettingValue: "3",
			Description:  "Berapa hari sebelum tenggat temuan sistem mengirim pengingat ke PIC & Manager (0 = nonaktif)",
		},
		{
			SettingKey:   master.SettingKeyNotifyOnKawasanComplete,
			SettingValue: "true",
			Description:  "Kirim email ke PIC & Manager saat seluruh Detail Kawasan di satu Kawasan selesai diinspeksi bulan ini",
		},
		{
			SettingKey:   master.SettingKeyNotifyOnAreaComplete,
			SettingValue: "true",
			Description:  "Kirim email ke PIC & Manager saat seluruh Kawasan di satu Area selesai diinspeksi",
		},
		{
			SettingKey:   master.SettingKeyIssueDeadlineDays,
			SettingValue: "14",
			Description:  "Batas waktu penyelesaian issue (dalam hari) sebelum menjadi overdue",
		},
		{
			SettingKey:   master.SettingKeyIssueAutoApproveDays,
			SettingValue: "3",
			Description:  "Batas waktu (hari) bagi auditor untuk memvalidasi follow up perbaikan temuan sebelum auto-approve",
		},
		{
			SettingKey:   master.SettingKeyIssueWOWRAutoApproveDays,
			SettingValue: "3",
			Description:  "Batas waktu (hari) bagi auditor untuk memvalidasi bukti WO/WR sebelum auto-approve",
		},
		{
			SettingKey:   master.SettingKeySMTPEnabled,
			SettingValue: "true",
			Description:  "Aktifkan atau nonaktifkan pengiriman email SMTP",
		},
		{
			SettingKey:  master.SettingKeySMTPHost,
			Description: "Hostname server SMTP; kosong menggunakan environment",
		},
		{
			SettingKey:  master.SettingKeySMTPPort,
			Description: "Port server SMTP; kosong menggunakan environment",
		},
		{
			SettingKey:  master.SettingKeySMTPUser,
			Description: "Username autentikasi SMTP; kosong menggunakan environment",
		},
		{
			SettingKey:  master.SettingKeySMTPPassword,
			IsEncrypted: true,
			Description: "Password atau app password SMTP (terenkripsi)",
		},
		{
			SettingKey:  master.SettingKeySMTPSenderEmail,
			Description: "Alamat email pengirim; kosong menggunakan environment",
		},
	}

	for _, setting := range settings {
		err := db.Where(&master.Setting{SettingKey: setting.SettingKey}).FirstOrCreate(&setting).Error
		if err != nil {
			log.Printf("⚠️ Failed to seed setting %s: %v", setting.SettingKey, err)
		}
	}

	// Update existing MINIO_ALLOWED_IPS setting to '*' for public read access over ngrok/docker
	_ = db.Model(&master.Setting{}).Where("\"SettingKey\" = ?", master.SettingKeyMinioAllowedIPs).Update("SettingValue", "*").Error

	log.Println("✅ Settings seeded successfully.")
}

package seeds

import (
	"log"

	"github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedSettings populates initial system settings
func SeedSettings(db *gorm.DB) {
	settings := []master.Setting{
		{
			SettingKey:   master.SettingKeyMinioAllowedIPs,
			SettingValue: "127.0.0.1,192.168.1.1",
			Description:  "Daftar IP yang diizinkan untuk bypass URL MinIO (pisahkan dengan koma)",
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
			SettingValue: "<h1>Reset Password</h1><p>Klik tombol berikut untuk mereset password Anda.</p>",
			Description:  "Template email untuk lupa password (HTML)",
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
			SettingKey:   master.SettingKeyIssueDeadlineDays,
			SettingValue: "14",
			Description:  "Batas waktu penyelesaian issue (dalam hari) sebelum menjadi overdue",
		},
		{
			SettingKey:   master.SettingKeyIssueAutoApproveDays,
			SettingValue: "3",
			Description:  "Batas waktu (hari) bagi auditor untuk memvalidasi follow up PIC sebelum auto-approve",
		},
	}

	for _, setting := range settings {
		err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "SettingKey"}},
			DoNothing: true,
		}).Create(&setting).Error
		if err != nil {
			log.Fatalf("❌ Failed to seed setting %s: %v", setting.SettingKey, err)
		}
	}

	log.Println("✅ Settings seeded successfully.")
}

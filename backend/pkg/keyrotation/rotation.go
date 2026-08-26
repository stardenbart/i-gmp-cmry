package keyrotation

import (
	"fmt"

	appcrypto "github.com/monitoring-system/backend/pkg/crypto"
	"gorm.io/gorm"
)

// Stats contains only record counts; rotation never returns plaintext values.
type Stats struct {
	SettingsRotated  int
	IssuesRotated    int
	PhotosRotated    int
	AlreadyCurrent   int
	PlaintextSkipped int
}

type settingRow struct {
	SettingKey   string `gorm:"column:SettingKey"`
	PlantID      string `gorm:"column:PlantID"`
	SettingValue string `gorm:"column:SettingValue"`
}

type issueRow struct {
	IssueID    string `gorm:"column:IssueID"`
	Keterangan string `gorm:"column:Keterangan"`
}

type photoRow struct {
	IssuePhotoID string `gorm:"column:IssuePhotoID"`
	ImageURL     string `gorm:"column:ImageUrl"`
	FileName     string `gorm:"column:FileName"`
	Keterangan   string `gorm:"column:Keterangan"`
}

// Rotate re-encrypts legacy ciphertext using newKey. When apply is false it
// validates every candidate and reports counts without changing the database.
func Rotate(db *gorm.DB, oldKey, newKey *appcrypto.Service, apply bool) (Stats, error) {
	var stats Stats
	run := func(tx *gorm.DB) error {
		if err := rotateSettings(tx, oldKey, newKey, apply, &stats); err != nil {
			return err
		}
		if err := rotateIssues(tx, oldKey, newKey, apply, &stats); err != nil {
			return err
		}
		return rotatePhotos(tx, oldKey, newKey, apply, &stats)
	}
	if apply {
		if err := db.Transaction(run); err != nil {
			return Stats{}, err
		}
		return stats, nil
	}
	if err := run(db); err != nil {
		return Stats{}, err
	}
	return stats, nil
}

func rotateValue(value string, required bool, oldKey, newKey *appcrypto.Service) (string, string, error) {
	if value == "" {
		return value, "empty", nil
	}
	if appcrypto.IsVersionedCiphertext(value) {
		if _, err := newKey.Decrypt(value); err == nil {
			return value, "current", nil
		}
	}
	plaintext, err := oldKey.Decrypt(value)
	if err != nil {
		if required || appcrypto.IsVersionedCiphertext(value) {
			return "", "", fmt.Errorf("value cannot be decrypted with old or new key")
		}
		return value, "plaintext", nil
	}
	rotated, err := newKey.Encrypt(plaintext)
	if err != nil {
		return "", "", err
	}
	return rotated, "rotated", nil
}

func rotateSettings(db *gorm.DB, oldKey, newKey *appcrypto.Service, apply bool, stats *Stats) error {
	var rows []settingRow
	if err := db.Table("System_Setting").Select("\"SettingKey\", \"PlantID\", \"SettingValue\"").Where("\"IsEncrypted\" = ? AND \"SettingValue\" <> ''", true).Find(&rows).Error; err != nil {
		return fmt.Errorf("read encrypted settings: %w", err)
	}
	for _, row := range rows {
		value, state, err := rotateValue(row.SettingValue, true, oldKey, newKey)
		if err != nil {
			return fmt.Errorf("setting %s/%s: %w", row.SettingKey, row.PlantID, err)
		}
		if state == "current" {
			stats.AlreadyCurrent++
			continue
		}
		stats.SettingsRotated++
		if apply {
			if err := db.Table("System_Setting").Where("\"SettingKey\" = ? AND \"PlantID\" = ?", row.SettingKey, row.PlantID).Update("SettingValue", value).Error; err != nil {
				return fmt.Errorf("update setting %s/%s: %w", row.SettingKey, row.PlantID, err)
			}
		}
	}
	return nil
}

func rotateIssues(db *gorm.DB, oldKey, newKey *appcrypto.Service, apply bool, stats *Stats) error {
	var rows []issueRow
	if err := db.Table("Issue").Select("\"IssueID\", \"Keterangan\"").Where("\"Keterangan\" IS NOT NULL AND \"Keterangan\" <> ''").Find(&rows).Error; err != nil {
		return fmt.Errorf("read issues: %w", err)
	}
	for _, row := range rows {
		value, state, err := rotateValue(row.Keterangan, false, oldKey, newKey)
		if err != nil {
			return fmt.Errorf("issue %s: %w", row.IssueID, err)
		}
		switch state {
		case "current":
			stats.AlreadyCurrent++
		case "plaintext":
			stats.PlaintextSkipped++
		case "rotated":
			stats.IssuesRotated++
			if apply {
				if err := db.Table("Issue").Where("\"IssueID\" = ?", row.IssueID).Update("Keterangan", value).Error; err != nil {
					return fmt.Errorf("update issue %s: %w", row.IssueID, err)
				}
			}
		}
	}
	return nil
}

func rotatePhotos(db *gorm.DB, oldKey, newKey *appcrypto.Service, apply bool, stats *Stats) error {
	var rows []photoRow
	if err := db.Table("Issue_Photo").Select("\"IssuePhotoID\", \"ImageUrl\", \"FileName\", \"Keterangan\"").Find(&rows).Error; err != nil {
		return fmt.Errorf("read issue photos: %w", err)
	}
	for _, row := range rows {
		updates := map[string]any{}
		for column, original := range map[string]string{"ImageUrl": row.ImageURL, "FileName": row.FileName, "Keterangan": row.Keterangan} {
			value, state, err := rotateValue(original, false, oldKey, newKey)
			if err != nil {
				return fmt.Errorf("issue photo %s column %s: %w", row.IssuePhotoID, column, err)
			}
			switch state {
			case "current":
				stats.AlreadyCurrent++
			case "plaintext":
				stats.PlaintextSkipped++
			case "rotated":
				updates[column] = value
			}
		}
		if len(updates) > 0 {
			stats.PhotosRotated += len(updates)
			if apply {
				if err := db.Table("Issue_Photo").Where("\"IssuePhotoID\" = ?", row.IssuePhotoID).Updates(updates).Error; err != nil {
					return fmt.Errorf("update issue photo %s: %w", row.IssuePhotoID, err)
				}
			}
		}
	}
	return nil
}

package migrator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// RunMigrations executes all unapplied SQL migration files in migrationsDir.
func RunMigrations(db *gorm.DB, migrationsDir string) error {
	// Create schema_migrations table if not exists
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		// Fallback to relative path if backend folder context
		if os.IsNotExist(err) {
			migrationsDir = "../migrations"
			files, err = os.ReadDir(migrationsDir)
		}
		if err != nil {
			log.Printf("⚠️  Migrations directory not found (%s), skipping SQL file migrations", migrationsDir)
			return nil
		}
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") && !strings.HasSuffix(f.Name(), "_down.sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, file := range sqlFiles {
		var count int64
		db.Table("schema_migrations").Where("version = ?", file).Count(&count)
		if count > 0 {
			continue // Already applied
		}

		log.Printf("Applying migration: %s", file)
		content, readErr := os.ReadFile(filepath.Join(migrationsDir, file))
		if readErr != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, readErr)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if execErr := tx.Exec(string(content)).Error; execErr != nil {
				return fmt.Errorf("failed to execute migration %s: %w", file, execErr)
			}
			return tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", file).Error
		})

		if err != nil {
			return fmt.Errorf("migration failed for %s: %w", file, err)
		}
		log.Printf("✅ Migration applied: %s", file)
	}

	return nil
}

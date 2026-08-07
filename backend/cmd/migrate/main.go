package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monitoring-system/backend/config"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect database: %v", err)
	}

	// Buat tabel history migrasi jika belum ada
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create migrations table: %v", err)
	}

	migrationsDir := "./migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("❌ Failed to read migrations directory: %v", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
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
		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			log.Fatalf("❌ Failed to read migration file %s: %v", file, err)
		}

		// Split statements and execute
		statements := strings.Split(string(content), ";")
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, stmt := range statements {
				stmt = strings.TrimSpace(stmt)
				if stmt == "" {
					continue
				}
				if err := tx.Exec(stmt).Error; err != nil {
					return fmt.Errorf("failed to execute statement in %s: %w\nStatement: %s", file, err, stmt)
				}
			}
			return tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", file).Error
		})

		if err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		log.Printf("✅ Migration applied: %s", file)
	}

	log.Println("🎉 All migrations applied successfully.")
}

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

		statements := splitSQLStatements(string(content))
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, stmt := range statements {
				stmt = strings.TrimSpace(stmt)
				if stmt == "" {
					continue
				}
				if execErr := tx.Exec(stmt).Error; execErr != nil {
					return fmt.Errorf("failed to execute statement in %s: %w\nStatement: %s", file, execErr, stmt)
				}
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

// splitSQLStatements splits a multi-statement SQL script into individual executable statements,
// respecting PL/pgSQL dollar-quoted blocks ($$).
func splitSQLStatements(script string) []string {
	var statements []string
	var current strings.Builder
	inDollarQuote := false

	lines := strings.Split(script, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Toggle dollar quote mode if line contains $$
		if strings.Contains(line, "$$") {
			inDollarQuote = !inDollarQuote
		}

		// Skip pure comment lines outside dollar quotes
		if !inDollarQuote && strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// End statement if not inside dollar quote and line ends with semicolon
		if !inDollarQuote && strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	remainder := strings.TrimSpace(current.String())
	if remainder != "" {
		statements = append(statements, remainder)
	}

	return statements
}

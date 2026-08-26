package seeds

import (
	"fmt"
	"log"

	"github.com/monitoring-system/backend/pkg/migrator"
	"gorm.io/gorm"
)

// Run executes all seeders in dependency order.
func Run(db *gorm.DB) error {
	log.Println("🌱 Starting database seeding...")

	// 1. Run SQL migrations to create all database tables
	if err := migrator.RunMigrations(db, "./migrations"); err != nil {
		return fmt.Errorf("run migrations before seeding: %w", err)
	}

	SeedRoles(db)
	SeedModules(db)
	SeedDepartments(db)
	SeedPlants(db)
	if err := SeedUsers(db); err != nil {
		return err
	}
	SeedPermissions(db)
	SeedSettings(db)
	SeedRealMasterData(db)
	CleanupDummyData(db)

	log.Println("✅ Seeding completed.")
	return nil
}

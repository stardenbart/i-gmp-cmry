package seeds

import (
	"log"

	"gorm.io/gorm"
)

// Run executes all seeders in dependency order.
func Run(db *gorm.DB) {
	log.Println("🌱 Starting database seeding...")

	SeedRoles(db)
	SeedModules(db)
	SeedPermissions(db)
	SeedDepartments(db)
	SeedPlants(db)
	SeedUsers(db)
	SeedSettings(db)
	CleanupDummyData(db)

	log.Println("✅ Seeding completed.")
}

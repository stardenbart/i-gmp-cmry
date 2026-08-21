package seeds

import (
	"log"

	inspectiondomain "github.com/monitoring-system/backend/internal/domain/inspection"
	issuedomain "github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
)

// Run executes all seeders in dependency order.
func Run(db *gorm.DB) {
	log.Println("🌱 Starting database seeding...")

	_ = db.AutoMigrate(
		&inspectiondomain.InspectionHeader{},
		&inspectiondomain.InspectionResult{},
		&issuedomain.Issue{},
		&issuedomain.IssuePhoto{},
		&issuedomain.IssueHEI{},
	)

	SeedRoles(db)
	SeedModules(db)
	SeedPermissions(db)
	SeedDepartments(db)
	SeedPlants(db)
	SeedUsers(db)
	SeedSettings(db)
	SeedRealMasterData(db)
	CleanupDummyData(db)

	log.Println("✅ Seeding completed.")
}

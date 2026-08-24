package seeds

import (
	"log"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	inspectiondomain "github.com/monitoring-system/backend/internal/domain/inspection"
	issuedomain "github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	picdomain "github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/migrator"
	"gorm.io/gorm"
)

// Run executes all seeders in dependency order.
func Run(db *gorm.DB) {
	log.Println("🌱 Starting database seeding...")

	// 1. Run SQL migrations to create all database tables
	if err := migrator.RunMigrations(db, "./migrations"); err != nil {
		log.Printf("⚠️  Migration warning: %v", err)
	}

	// AutoMigrate all domain tables to guarantee table existence on fresh database
	_ = db.AutoMigrate(
		&authdomain.Role{},
		&authdomain.Module{},
		&authdomain.Permission{},
		&authdomain.RolePermission{},
		&authdomain.UserPermission{},
		&authdomain.User{},
		&masterdomain.Plant{},
		&masterdomain.Department{},
		&masterdomain.Area{},
		&masterdomain.Kawasan{},
		&masterdomain.DetailKawasan{},
		&masterdomain.Aspek{},
		&masterdomain.Detail{},
		&masterdomain.Uraian{},
		&masterdomain.Setting{},
		&masterdomain.HEIMaster{},
		&picdomain.PICMapping{},
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

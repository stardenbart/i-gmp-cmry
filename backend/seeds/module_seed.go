package seeds

import (
	"log"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

// defaultModules maps to the main feature areas of the application.
var defaultModules = []authdomain.Module{
	{ModuleID: "MOD-USR", ModuleName: "User Management"},
	{ModuleID: "MOD-ROLE", ModuleName: "Role & Permission Management"},
	{ModuleID: "MOD-MSTR", ModuleName: "Master Data"},
	{ModuleID: "MOD-PIC", ModuleName: "PIC Mapping"},
	{ModuleID: "MOD-INSP", ModuleName: "Inspection"},
	{ModuleID: "MOD-ISS", ModuleName: "Issue & Follow Up"},
	{ModuleID: "MOD-LOG", ModuleName: "Logging & Audit Trail"},
}

// SeedModules inserts default application modules if they don't exist.
func SeedModules(db *gorm.DB) {
	for _, m := range defaultModules {
		var count int64
		db.Model(&authdomain.Module{}).Where(&authdomain.Module{ModuleID: m.ModuleID}).Count(&count)
		if count == 0 {
			if err := db.Create(&m).Error; err != nil {
				log.Printf("❌ Failed to seed module %s: %v", m.ModuleName, err)
			} else {
				log.Printf("   ✔ Module seeded: %s", m.ModuleName)
			}
		}
	}
}

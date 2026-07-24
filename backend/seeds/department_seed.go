package seeds

import (
	"log"

	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

var defaultDepartments = []masterdomain.Department{
	{DepartmentID: "DEPT-001", DepartmentName: "Quality Control (QC)"},
	{DepartmentID: "DEPT-002", DepartmentName: "Quality Assurance (QA)"},
	{DepartmentID: "DEPT-003", DepartmentName: "Produksi"},
	{DepartmentID: "DEPT-004", DepartmentName: "Maintenance"},
	{DepartmentID: "DEPT-005", DepartmentName: "Warehouse"},
	{DepartmentID: "DEPT-006", DepartmentName: "HSE (Health, Safety & Environment)"},
	{DepartmentID: "DEPT-007", DepartmentName: "IT"},
}

// SeedDepartments inserts default departments if they don't exist.
func SeedDepartments(db *gorm.DB) {
	for _, d := range defaultDepartments {
		var count int64
		db.Model(&masterdomain.Department{}).Where(&masterdomain.Department{DepartmentID: d.DepartmentID}).Count(&count)
		if count == 0 {
			if err := db.Create(&d).Error; err != nil {
				log.Printf("Failed to seed department %s: %v", d.DepartmentName, err)
			} else {
				log.Printf("Department seeded: %s", d.DepartmentName)
			}
		}
	}
}

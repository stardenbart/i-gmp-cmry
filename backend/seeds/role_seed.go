package seeds

import (
	"log"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

var defaultRoles = []authdomain.Role{
	{RoleID: "ROLE-000", RoleName: "Super Admin", RoleDescription: "Super Administrator sistem dengan akses penuh global"},
	{RoleID: "ROLE-001", RoleName: "Admin", RoleDescription: "Administrator sistem dengan akses di plant terkait"},
	{RoleID: "ROLE-002", RoleName: "Auditor", RoleDescription: "Petugas yang melakukan inspeksi/audit"},
	{RoleID: "ROLE-003", RoleName: "Auditee", RoleDescription: "Pihak yang diaudit, PIC area/kawasan"},
	{RoleID: "ROLE-004", RoleName: "Supervisor", RoleDescription: "Supervisor yang memantau proses audit"},
	{RoleID: "ROLE-005", RoleName: "Manager", RoleDescription: "Manajer dengan akses laporan dan approval"},
	{RoleID: "ROLE-006", RoleName: "Staff", RoleDescription: "Staff operasional lapangan"},
}

// SeedRoles inserts default roles if they don't exist.
func SeedRoles(db *gorm.DB) {
	for _, role := range defaultRoles {
		var count int64
		db.Model(&authdomain.Role{}).Where(&authdomain.Role{RoleID: role.RoleID}).Count(&count)
		if count == 0 {
			if err := db.Create(&role).Error; err != nil {
				log.Printf("❌ Failed to seed role %s: %v", role.RoleName, err)
			} else {
				log.Printf("   ✔ Role seeded: %s", role.RoleName)
			}
		}
	}
}

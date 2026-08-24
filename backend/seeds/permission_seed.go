package seeds

import (
	"fmt"
	"log"
	"strings"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

// defaultPermissions defines all CRUD/action permissions per module.
var defaultPermissions = []authdomain.Permission{
	// User Management
	{PermissionID: "PERM-USR-C", ModuleID: "MOD-USR", PermissionCode: "CREATE", PermissionName: "Create User"},
	{PermissionID: "PERM-USR-R", ModuleID: "MOD-USR", PermissionCode: "READ", PermissionName: "View Users"},
	{PermissionID: "PERM-USR-U", ModuleID: "MOD-USR", PermissionCode: "UPDATE", PermissionName: "Update User"},
	{PermissionID: "PERM-USR-D", ModuleID: "MOD-USR", PermissionCode: "DELETE", PermissionName: "Delete User"},

	// Role & Permission
	{PermissionID: "PERM-ROLE-R", ModuleID: "MOD-ROLE", PermissionCode: "READ", PermissionName: "View Roles"},
	{PermissionID: "PERM-ROLE-U", ModuleID: "MOD-ROLE", PermissionCode: "UPDATE", PermissionName: "Manage Role Permissions"},

	// Master Data
	{PermissionID: "PERM-MSTR-C", ModuleID: "MOD-MSTR", PermissionCode: "CREATE", PermissionName: "Create Master Data"},
	{PermissionID: "PERM-MSTR-R", ModuleID: "MOD-MSTR", PermissionCode: "READ", PermissionName: "View Master Data"},
	{PermissionID: "PERM-MSTR-U", ModuleID: "MOD-MSTR", PermissionCode: "UPDATE", PermissionName: "Update Master Data"},
	{PermissionID: "PERM-MSTR-D", ModuleID: "MOD-MSTR", PermissionCode: "DELETE", PermissionName: "Delete Master Data"},

	// PIC Mapping
	{PermissionID: "PERM-PIC-C", ModuleID: "MOD-PIC", PermissionCode: "CREATE", PermissionName: "Create PIC Mapping"},
	{PermissionID: "PERM-PIC-R", ModuleID: "MOD-PIC", PermissionCode: "READ", PermissionName: "View PIC Mapping"},
	{PermissionID: "PERM-PIC-U", ModuleID: "MOD-PIC", PermissionCode: "UPDATE", PermissionName: "Update PIC Mapping"},
	{PermissionID: "PERM-PIC-D", ModuleID: "MOD-PIC", PermissionCode: "DELETE", PermissionName: "Delete PIC Mapping"},

	// Inspection
	{PermissionID: "PERM-INSP-C", ModuleID: "MOD-INSP", PermissionCode: "CREATE", PermissionName: "Create Inspection"},
	{PermissionID: "PERM-INSP-R", ModuleID: "MOD-INSP", PermissionCode: "READ", PermissionName: "View Inspections"},
	{PermissionID: "PERM-INSP-U", ModuleID: "MOD-INSP", PermissionCode: "UPDATE", PermissionName: "Update Inspection"},
	{PermissionID: "PERM-INSP-A", ModuleID: "MOD-INSP", PermissionCode: "APPROVE", PermissionName: "Approve Inspection"},
	{PermissionID: "PERM-INSP-E", ModuleID: "MOD-INSP", PermissionCode: "EXPORT", PermissionName: "Export Inspection Report"},

	// Issue & Follow Up
	{PermissionID: "PERM-ISS-C", ModuleID: "MOD-ISS", PermissionCode: "CREATE", PermissionName: "Create Issue"},
	{PermissionID: "PERM-ISS-R", ModuleID: "MOD-ISS", PermissionCode: "READ", PermissionName: "View Issues"},
	{PermissionID: "PERM-ISS-U", ModuleID: "MOD-ISS", PermissionCode: "UPDATE", PermissionName: "Update Issue / Follow Up"},

	// Logging
	{PermissionID: "PERM-LOG-R", ModuleID: "MOD-LOG", PermissionCode: "READ", PermissionName: "View Logs"},

	// Data Inspeksi (GMP)
	{PermissionID: "PERM-GMP-R", ModuleID: "MOD-GMP", PermissionCode: "READ", PermissionName: "View GMP Data"},

	// Pengaturan Sistem
	{PermissionID: "PERM-STNG-R", ModuleID: "MOD-STNG", PermissionCode: "READ", PermissionName: "View Settings"},
	{PermissionID: "PERM-STNG-U", ModuleID: "MOD-STNG", PermissionCode: "UPDATE", PermissionName: "Update Settings"},

	// Perintah Kerja (WO/WR)
	{PermissionID: "PERM-WOWR-R", ModuleID: "MOD-WOWR", PermissionCode: "READ", PermissionName: "View Perintah Kerja (WO/WR)"},
	{PermissionID: "PERM-WOWR-U", ModuleID: "MOD-WOWR", PermissionCode: "UPDATE", PermissionName: "Update / Validasi Bukti WO/WR"},
}

// SeedPermissions inserts all module permissions and grants default permissions to Admin in Role_Permission table.
func SeedPermissions(db *gorm.DB) {
	for _, p := range defaultPermissions {
		var count int64
		db.Model(&authdomain.Permission{}).Where(&authdomain.Permission{PermissionID: p.PermissionID}).Count(&count)
		if count == 0 {
			if err := db.Create(&p).Error; err != nil {
				log.Printf("❌ Failed to seed permission %s: %v", p.PermissionName, err)
			} else {
				log.Printf("   ✔ Permission seeded: %s", p.PermissionName)
			}
		}

		// Ensure ROLE-001 (Admin) has explicit true entries in Role_Permission table
		var rp authdomain.RolePermission
		rpID := "RP-001-" + strings.TrimPrefix(p.PermissionID, "PERM-")
		if err := db.Where("\"RoleID\" = ? AND \"PermissionID\" = ?", "ROLE-001", p.PermissionID).First(&rp).Error; err != nil {
			_ = db.Create(&authdomain.RolePermission{
				RolePermissionID:        rpID,
				RoleID:                  "ROLE-001",
				PermissionID:            p.PermissionID,
				IsAllowed:               true,
				RolePermissionUpdatedBy: "USR-ADMIN-001",
			})
		} else if !rp.IsAllowed {
			_ = db.Model(&rp).Update("IsAllowed", true).Error
		}
	}

	// Seed default permissions for Non-Admin Roles (Auditor, Auditee, Supervisor, Manager, Staff)
	// Non-Admin roles receive PERM-MSTR-R so they can view Master Data dropdowns (Area, Department, Kawasan, etc.)
	defaultRolePerms := map[string][]string{
		"ROLE-002": {"PERM-MSTR-R", "PERM-INSP-C", "PERM-INSP-R", "PERM-INSP-U", "PERM-INSP-A", "PERM-INSP-E", "PERM-ISS-C", "PERM-ISS-R", "PERM-ISS-U", "PERM-WOWR-R", "PERM-WOWR-U"}, // Auditor
		"ROLE-003": {"PERM-MSTR-R", "PERM-ISS-R", "PERM-ISS-U", "PERM-WOWR-R", "PERM-WOWR-U"},                                                                                       // Auditee
		"ROLE-004": {"PERM-MSTR-R", "PERM-INSP-R", "PERM-ISS-R", "PERM-WOWR-R"},                                                                                                     // Supervisor
		"ROLE-005": {"PERM-MSTR-R", "PERM-INSP-R", "PERM-INSP-E", "PERM-ISS-R", "PERM-WOWR-R"},                                                                                       // Manager
		"ROLE-006": {"PERM-MSTR-R", "PERM-ISS-R", "PERM-ISS-U", "PERM-WOWR-R"},                                                                                                       // Staff
	}

	for roleID, permIDs := range defaultRolePerms {
		roleNum := strings.TrimPrefix(roleID, "ROLE-")
		for _, permID := range permIDs {
			var rp authdomain.RolePermission
			rpID := fmt.Sprintf("RP-%s-%s", roleNum, strings.TrimPrefix(permID, "PERM-"))
			if err := db.Where("\"RoleID\" = ? AND \"PermissionID\" = ?", roleID, permID).First(&rp).Error; err != nil {
				_ = db.Create(&authdomain.RolePermission{
					RolePermissionID:        rpID,
					RoleID:                  roleID,
					PermissionID:            permID,
					IsAllowed:               true,
					RolePermissionUpdatedBy: "SYSTEM",
				})
			} else if !rp.IsAllowed {
				_ = db.Model(&rp).Update("IsAllowed", true).Error
			}
		}
	}
}

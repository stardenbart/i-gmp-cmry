package seeds

import (
	"log"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/password"
	"gorm.io/gorm"
)

type SeedUserDef struct {
	UserID       string
	DepartmentID string
	RoleID       string
	PlantID      *string
	Username     string
	FullName     string
	Email        string
	Password     string
}

func strPtr(s string) *string {
	return &s
}

var defaultUsers = []SeedUserDef{
	{
		UserID:       "USR-SUPERADMIN-001",
		DepartmentID: "DEPT-007",
		RoleID:       "ROLE-000", // Super Admin
		PlantID:      nil,        // Global
		Username:     "superadmin",
		FullName:     "Global Super Admin",
		Email:        "superadmin@cimory.com",
		Password:     "admin123",
	},
	{
		UserID:       "USR-ADMIN-001",
		DepartmentID: "DEPT-007",
		RoleID:       "ROLE-001", // Admin
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "admin",
		FullName:     "Admin Sentul",
		Email:        "admin@cimory.com",
		Password:     "admin123",
	},
	{
		UserID:       "USR-AUDIT-001",
		DepartmentID: "DEPT-001", // Quality Assurance
		RoleID:       "ROLE-002", // Auditor
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditor",
		FullName:     "Inspektur Auditor",
		Email:        "auditor@cimory.com",
		Password:     "auditor123",
	},
	{
		UserID:       "USR-AUDITEE-001",
		DepartmentID: "DEPT-002", // Production
		RoleID:       "ROLE-003", // Auditee
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditee",
		FullName:     "PIC Auditee",
		Email:        "auditee@cimory.com",
		Password:     "auditee123",
	},
	{
		UserID:       "USR-SUPERVISOR-001",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-004", // Supervisor
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "supervisor",
		FullName:     "Supervisor Area",
		Email:        "supervisor@cimory.com",
		Password:     "supervisor123",
	},
	{
		UserID:       "USR-MANAGER-001",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-005", // Manager
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "manager",
		FullName:     "Manager QA",
		Email:        "manager@cimory.com",
		Password:     "manager123",
	},
	{
		UserID:       "USR-STAFF-001",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-006", // Staff
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "staff",
		FullName:     "Staff Operasional",
		Email:        "staff@cimory.com",
		Password:     "staff123",
	},
}

// SeedUsers creates default users for all roles.
func SeedUsers(db *gorm.DB) {
	for _, u := range defaultUsers {
		var count int64
		db.Model(&authdomain.User{}).Where("username = ?", u.Username).Count(&count)
		if count == 0 {
			hashed, err := password.Hash(u.Password)
			if err != nil {
				log.Printf("Failed to hash password for %s: %v", u.Username, err)
				continue
			}

			user := authdomain.User{
				UserID:       u.UserID,
				DepartmentID: u.DepartmentID,
				RoleID:       u.RoleID,
				PlantID:      u.PlantID,
				Username:     u.Username,
				FullName:     u.FullName,
				Email:        u.Email,
				PasswordHash: hashed,
				UserStatus:   authdomain.UserStatusActive,
			}

			if err := db.Create(&user).Error; err != nil {
				log.Printf("❌ Failed to seed user %s: %v", u.Username, err)
			} else {
				log.Printf("   ✔ User seeded (username: %s, password: %s)", u.Username, u.Password)
			}
		} else {
			// Update PlantID if missing on existing user
			if u.PlantID != nil {
				_ = db.Model(&authdomain.User{}).Where("username = ? AND (\"PlantID\" IS NULL OR \"PlantID\" = '')", u.Username).Update("PlantID", u.PlantID).Error
			}
		}
	}
}

// Backwards compatibility wrappers
func SeedAdminUser(db *gorm.DB)   { SeedUsers(db) }
func SeedAuditorUser(db *gorm.DB) {}
func SeedAuditeeUser(db *gorm.DB) {}

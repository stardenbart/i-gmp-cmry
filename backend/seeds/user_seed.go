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
	// ─── 0. GLOBAL (SUPER ADMIN) ──────────────────────────────────────────────
	{
		UserID:       "USR-SUPERADMIN-001",
		DepartmentID: "DEPT-007",
		RoleID:       "ROLE-000", // Super Admin
		PlantID:      nil,        // Global Scope
		Username:     "superadmin",
		FullName:     "Global Super Admin",
		Email:        "superadmin@cimory.com",
		Password:     "admin123",
	},

	// ─── 1. PLANT SENTUL (PLT-SENTUL) ─────────────────────────────────────────
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
		UserID:       "USR-ADMIN-SENTUL",
		DepartmentID: "DEPT-007",
		RoleID:       "ROLE-001",
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "admin_sentul",
		FullName:     "Admin Plant Sentul",
		Email:        "admin.sentul@cimory.com",
		Password:     "admin123",
	},
	{
		UserID:       "USR-AUDIT-001",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-002", // Auditor
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditor",
		FullName:     "Inspektur Auditor Sentul",
		Email:        "auditor@cimory.com",
		Password:     "auditor123",
	},
	{
		UserID:       "USR-AUDIT-SENTUL",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-002",
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditor_sentul",
		FullName:     "Auditor Plant Sentul",
		Email:        "auditor.sentul@cimory.com",
		Password:     "auditor123",
	},
	{
		UserID:       "USR-AUDITEE-001",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-003", // Auditee
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditee",
		FullName:     "PIC Auditee Sentul",
		Email:        "auditee@cimory.com",
		Password:     "auditee123",
	},
	{
		UserID:       "USR-AUDITEE-SENTUL",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-003",
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditee_sentul",
		FullName:     "PIC Auditee Plant Sentul",
		Email:        "auditee.sentul@cimory.com",
		Password:     "auditee123",
	},
	{
		UserID:       "USR-SUPERVISOR-001",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-004", // Supervisor
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "supervisor",
		FullName:     "Supervisor Sentul",
		Email:        "supervisor@cimory.com",
		Password:     "supervisor123",
	},
	{
		UserID:       "USR-MANAGER-001",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-005", // Manager
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "manager",
		FullName:     "Manager QA Sentul",
		Email:        "manager@cimory.com",
		Password:     "manager123",
	},
	{
		UserID:       "USR-STAFF-001",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-006", // Staff
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "staff",
		FullName:     "Staff Operasional Sentul",
		Email:        "staff@cimory.com",
		Password:     "staff123",
	},

	// ─── 2. PLANT CICURUG (PLT-CICURUG) ───────────────────────────────────────
	{
		UserID:       "USR-ADMIN-CICURUG",
		DepartmentID: "DEPT-007",
		RoleID:       "ROLE-001", // Admin
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "admin_cicurug",
		FullName:     "Admin Plant Cicurug",
		Email:        "admin.cicurug@cimory.com",
		Password:     "admin123",
	},
	{
		UserID:       "USR-AUDIT-CICURUG",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-002", // Auditor
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "auditor_cicurug",
		FullName:     "Auditor Plant Cicurug",
		Email:        "auditor.cicurug@cimory.com",
		Password:     "auditor123",
	},
	{
		UserID:       "USR-AUDITEE-CICURUG",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-003", // Auditee
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "auditee_cicurug",
		FullName:     "PIC Auditee Plant Cicurug",
		Email:        "auditee.cicurug@cimory.com",
		Password:     "auditee123",
	},
	{
		UserID:       "USR-SUPERVISOR-CICURUG",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-004", // Supervisor
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "supervisor_cicurug",
		FullName:     "Supervisor Plant Cicurug",
		Email:        "supervisor.cicurug@cimory.com",
		Password:     "supervisor123",
	},
	{
		UserID:       "USR-MANAGER-CICURUG",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-005", // Manager
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "manager_cicurug",
		FullName:     "Manager QA Plant Cicurug",
		Email:        "manager.cicurug@cimory.com",
		Password:     "manager123",
	},
	{
		UserID:       "USR-STAFF-CICURUG",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-006", // Staff
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "staff_cicurug",
		FullName:     "Staff Operasional Cicurug",
		Email:        "staff.cicurug@cimory.com",
		Password:     "staff123",
	},

	// ─── 3. PLANT PASURUAN (PLT-PASURUAN) ────────────────────────────────────
	{
		UserID:       "USR-ADMIN-PASURUAN",
		DepartmentID: "DEPT-007",
		RoleID:       "ROLE-001", // Admin
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "admin_pasuruan",
		FullName:     "Admin Plant Pasuruan",
		Email:        "admin.pasuruan@cimory.com",
		Password:     "admin123",
	},
	{
		UserID:       "USR-AUDIT-PASURUAN",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-002", // Auditor
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "auditor_pasuruan",
		FullName:     "Auditor Plant Pasuruan",
		Email:        "auditor.pasuruan@cimory.com",
		Password:     "auditor123",
	},
	{
		UserID:       "USR-AUDITEE-PASURUAN",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-003", // Auditee
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "auditee_pasuruan",
		FullName:     "PIC Auditee Plant Pasuruan",
		Email:        "auditee.pasuruan@cimory.com",
		Password:     "auditee123",
	},
	{
		UserID:       "USR-SUPERVISOR-PASURUAN",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-004", // Supervisor
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "supervisor_pasuruan",
		FullName:     "Supervisor Plant Pasuruan",
		Email:        "supervisor.pasuruan@cimory.com",
		Password:     "supervisor123",
	},
	{
		UserID:       "USR-MANAGER-PASURUAN",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-005", // Manager
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "manager_pasuruan",
		FullName:     "Manager QA Plant Pasuruan",
		Email:        "manager.pasuruan@cimory.com",
		Password:     "manager123",
	},
	{
		UserID:       "USR-STAFF-PASURUAN",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-006", // Staff
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "staff_pasuruan",
		FullName:     "Staff Operasional Pasuruan",
		Email:        "staff.pasuruan@cimory.com",
		Password:     "staff123",
	},
}

// SeedUsers creates default users for all roles across all plants.
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
				log.Printf("   ✔ User seeded (username: %s, password: %s, plant: %v)", u.Username, u.Password, u.PlantID)
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

package seeds

import (
	"fmt"
	"log"
	"os"

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
	},
	{
		UserID:       "USR-ADMIN-SENTUL",
		DepartmentID: "DEPT-007",
		RoleID:       "ROLE-001",
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "admin_sentul",
		FullName:     "Admin Plant Sentul",
		Email:        "admin.sentul@cimory.com",
	},
	{
		UserID:       "USR-AUDIT-001",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-002", // Auditor
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditor",
		FullName:     "Inspektur Auditor Sentul",
		Email:        "auditor@cimory.com",
	},
	{
		UserID:       "USR-AUDIT-SENTUL",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-002",
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditor_sentul",
		FullName:     "Auditor Plant Sentul",
		Email:        "auditor.sentul@cimory.com",
	},
	{
		UserID:       "USR-AUDITEE-001",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-003", // Auditee
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditee",
		FullName:     "PIC Auditee Sentul",
		Email:        "auditee@cimory.com",
	},
	{
		UserID:       "USR-AUDITEE-SENTUL",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-003",
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "auditee_sentul",
		FullName:     "PIC Auditee Plant Sentul",
		Email:        "auditee.sentul@cimory.com",
	},
	{
		UserID:       "USR-SUPERVISOR-001",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-004", // Supervisor
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "supervisor",
		FullName:     "Supervisor Sentul",
		Email:        "supervisor@cimory.com",
	},
	{
		UserID:       "USR-MANAGER-001",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-005", // Manager
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "manager",
		FullName:     "Manager QA Sentul",
		Email:        "manager@cimory.com",
	},
	{
		UserID:       "USR-STAFF-001",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-006", // Staff
		PlantID:      strPtr("PLT-SENTUL"),
		Username:     "staff",
		FullName:     "Staff Operasional Sentul",
		Email:        "staff@cimory.com",
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
	},
	{
		UserID:       "USR-AUDIT-CICURUG",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-002", // Auditor
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "auditor_cicurug",
		FullName:     "Auditor Plant Cicurug",
		Email:        "auditor.cicurug@cimory.com",
	},
	{
		UserID:       "USR-AUDITEE-CICURUG",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-003", // Auditee
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "auditee_cicurug",
		FullName:     "PIC Auditee Plant Cicurug",
		Email:        "auditee.cicurug@cimory.com",
	},
	{
		UserID:       "USR-SPV-CICURUG",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-004", // Supervisor
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "supervisor_cicurug",
		FullName:     "Supervisor Plant Cicurug",
		Email:        "supervisor.cicurug@cimory.com",
	},
	{
		UserID:       "USR-MANAGER-CICURUG",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-005", // Manager
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "manager_cicurug",
		FullName:     "Manager QA Plant Cicurug",
		Email:        "manager.cicurug@cimory.com",
	},
	{
		UserID:       "USR-STAFF-CICURUG",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-006", // Staff
		PlantID:      strPtr("PLT-CICURUG"),
		Username:     "staff_cicurug",
		FullName:     "Staff Operasional Cicurug",
		Email:        "staff.cicurug@cimory.com",
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
	},
	{
		UserID:       "USR-AUDIT-PASURUAN",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-002", // Auditor
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "auditor_pasuruan",
		FullName:     "Auditor Plant Pasuruan",
		Email:        "auditor.pasuruan@cimory.com",
	},
	{
		UserID:       "USR-AUDITEE-PASURUAN",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-003", // Auditee
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "auditee_pasuruan",
		FullName:     "PIC Auditee Plant Pasuruan",
		Email:        "auditee.pasuruan@cimory.com",
	},
	{
		UserID:       "USR-SPV-PASURUAN",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-004", // Supervisor
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "supervisor_pasuruan",
		FullName:     "Supervisor Plant Pasuruan",
		Email:        "supervisor.pasuruan@cimory.com",
	},
	{
		UserID:       "USR-MANAGER-PASURUAN",
		DepartmentID: "DEPT-001",
		RoleID:       "ROLE-005", // Manager
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "manager_pasuruan",
		FullName:     "Manager QA Plant Pasuruan",
		Email:        "manager.pasuruan@cimory.com",
	},
	{
		UserID:       "USR-STAFF-PASURUAN",
		DepartmentID: "DEPT-002",
		RoleID:       "ROLE-006", // Staff
		PlantID:      strPtr("PLT-PASURUAN"),
		Username:     "staff_pasuruan",
		FullName:     "Staff Operasional Pasuruan",
		Email:        "staff.pasuruan@cimory.com",
	},
}

func seedPassword(roleID string) (string, error) {
	envName := map[string]string{
		"ROLE-000": "SEED_ADMIN_PASSWORD",
		"ROLE-001": "SEED_ADMIN_PASSWORD",
		"ROLE-002": "SEED_AUDITOR_PASSWORD",
		"ROLE-003": "SEED_AUDITEE_PASSWORD",
		"ROLE-004": "SEED_SUPERVISOR_PASSWORD",
		"ROLE-005": "SEED_MANAGER_PASSWORD",
		"ROLE-006": "SEED_STAFF_PASSWORD",
	}[roleID]
	value := os.Getenv(envName)
	if envName == "" || len(value) < 32 {
		return "", fmt.Errorf("%s must be set to at least 32 characters", envName)
	}
	return value, nil
}

// SeedUsers creates default users for all roles across all plants.
func SeedUsers(db *gorm.DB) error {
	for _, u := range defaultUsers {
		var count int64
		db.Model(&authdomain.User{}).Where("\"Username\" = ?", u.Username).Count(&count)
		if count == 0 {
			plainPassword, err := seedPassword(u.RoleID)
			if err != nil {
				return err
			}
			hashed, err := password.Hash(plainPassword)
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
				log.Printf("   ✔ User seeded (username: %s, plant: %v)", u.Username, u.PlantID)
			}
		} else {
			// Update PlantID if missing on existing user
			if u.PlantID != nil {
				_ = db.Exec(`UPDATE "Users" SET "PlantID" = ? WHERE "Username" = ? AND ("PlantID" IS NULL OR "PlantID" = '')`, *u.PlantID, u.Username).Error
			}
		}
	}
	return nil
}

// RotateDefaultUserPasswords updates only the well-known accounts managed by
// this seeder. It never logs plaintext credentials and commits atomically.
func RotateDefaultUserPasswords(db *gorm.DB) (int64, error) {
	var updated int64
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, u := range defaultUsers {
			plainPassword, err := seedPassword(u.RoleID)
			if err != nil {
				return err
			}
			hashed, err := password.Hash(plainPassword)
			if err != nil {
				return fmt.Errorf("hash password for %s: %w", u.Username, err)
			}
			result := tx.Model(&authdomain.User{}).
				Where("\"UserID\" = ? AND \"Username\" = ?", u.UserID, u.Username).
				Update("PasswordHash", hashed)
			if result.Error != nil {
				return fmt.Errorf("update password for %s: %w", u.Username, result.Error)
			}
			updated += result.RowsAffected
		}
		return nil
	})
	return updated, err
}

// Backwards compatibility wrappers
func SeedAdminUser(db *gorm.DB)   { _ = SeedUsers(db) }
func SeedAuditorUser(db *gorm.DB) {}
func SeedAuditeeUser(db *gorm.DB) {}

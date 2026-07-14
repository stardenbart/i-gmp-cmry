package seeds

import (
	"log"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/password"
	"gorm.io/gorm"
)

// SeedAdminUser creates a default admin user if one doesn't exist.
func SeedAdminUser(db *gorm.DB) {
	var count int64
	db.Model(&authdomain.User{}).Where("Username = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hashed, err := password.Hash("admin123")
	if err != nil {
		log.Printf("❌ Failed to hash admin password: %v", err)
		return
	}

	admin := authdomain.User{
		UserID:       "USR-ADMIN-001",
		DepartmentID: "DEPT-007", // IT
		RoleID:       "ROLE-001", // Admin
		Username:     "admin",
		FullName:     "Super Admin",
		Email:        "admin@company.com",
		PasswordHash: hashed,
		UserStatus:   authdomain.UserStatusActive,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("❌ Failed to seed admin user: %v", err)
	} else {
		log.Printf("   ✔ Admin user seeded (username: admin, password: admin123)")
	}
}

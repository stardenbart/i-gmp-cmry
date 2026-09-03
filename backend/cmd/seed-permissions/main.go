// cmd/seed-permissions/main.go
// Jalankan dengan: go run cmd/seed-permissions/main.go
// Tool ini akan mengisi tabel Role_Permission dengan izin default untuk semua role.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/monitoring-system/backend/config"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
)

// rolePermissionEntry mendefinisikan satu entri izin untuk suatu role.
type rolePermissionEntry struct {
	id           string
	roleID       string
	permissionID string
}

// defaultRolePermissions mendefinisikan izin default per role sesuai desain sistem.
var defaultRolePermissions = []rolePermissionEntry{
	// ── ROLE-001: Admin (semua izin aktif) ──────────────────────────────────
	{"RP-ADM-USR-C", "ROLE-001", "PERM-USR-C"},
	{"RP-ADM-USR-R", "ROLE-001", "PERM-USR-R"},
	{"RP-ADM-USR-U", "ROLE-001", "PERM-USR-U"},
	{"RP-ADM-USR-D", "ROLE-001", "PERM-USR-D"},
	{"RP-ADM-ROLE-R", "ROLE-001", "PERM-ROLE-R"},
	{"RP-ADM-ROLE-U", "ROLE-001", "PERM-ROLE-U"},
	{"RP-ADM-MSTR-C", "ROLE-001", "PERM-MSTR-C"},
	{"RP-ADM-MSTR-R", "ROLE-001", "PERM-MSTR-R"},
	{"RP-ADM-MSTR-U", "ROLE-001", "PERM-MSTR-U"},
	{"RP-ADM-MSTR-D", "ROLE-001", "PERM-MSTR-D"},
	{"RP-ADM-MSTR-I", "ROLE-001", "PERM-MSTR-I"},
	{"RP-ADM-PIC-C", "ROLE-001", "PERM-PIC-C"},
	{"RP-ADM-PIC-R", "ROLE-001", "PERM-PIC-R"},
	{"RP-ADM-PIC-U", "ROLE-001", "PERM-PIC-U"},
	{"RP-ADM-PIC-D", "ROLE-001", "PERM-PIC-D"},
	{"RP-ADM-INSP-C", "ROLE-001", "PERM-INSP-C"},
	{"RP-ADM-INSP-R", "ROLE-001", "PERM-INSP-R"},
	{"RP-ADM-INSP-U", "ROLE-001", "PERM-INSP-U"},
	{"RP-ADM-INSP-A", "ROLE-001", "PERM-INSP-A"},
	{"RP-ADM-INSP-E", "ROLE-001", "PERM-INSP-E"},
	{"RP-ADM-ISS-C", "ROLE-001", "PERM-ISS-C"},
	{"RP-ADM-ISS-R", "ROLE-001", "PERM-ISS-R"},
	{"RP-ADM-ISS-U", "ROLE-001", "PERM-ISS-U"},
	{"RP-ADM-ISS-E", "ROLE-001", "PERM-ISS-E"},
	{"RP-ADM-LOG-R", "ROLE-001", "PERM-LOG-R"},

	// ── ROLE-002: Auditor (inspeksi penuh, issue update/wowr, log baca) ──────
	{"RP-AUD-INSP-C", "ROLE-002", "PERM-INSP-C"},
	{"RP-AUD-INSP-R", "ROLE-002", "PERM-INSP-R"},
	{"RP-AUD-INSP-U", "ROLE-002", "PERM-INSP-U"},
	{"RP-AUD-INSP-A", "ROLE-002", "PERM-INSP-A"},
	{"RP-AUD-INSP-E", "ROLE-002", "PERM-INSP-E"},
	{"RP-AUD-ISS-U", "ROLE-002", "PERM-ISS-U"},
	{"RP-AUD-ISS-E", "ROLE-002", "PERM-ISS-E"},
	{"RP-AUD-LOG-R", "ROLE-002", "PERM-LOG-R"},
	{"RP-AUD-MSTR-R", "ROLE-002", "PERM-MSTR-R"},
	{"RP-AUD-PIC-R", "ROLE-002", "PERM-PIC-R"},

	// ── ROLE-003: Auditee (hanya issue update/baca - submit WO/WR) ──────────
	{"RP-AUE-ISS-R", "ROLE-003", "PERM-ISS-R"},
	{"RP-AUE-ISS-U", "ROLE-003", "PERM-ISS-U"},
	{"RP-AUE-MSTR-R", "ROLE-003", "PERM-MSTR-R"},

	// ── ROLE-004: Supervisor (baca semua, approve inspeksi) ─────────────────
	{"RP-SUP-INSP-R", "ROLE-004", "PERM-INSP-R"},
	{"RP-SUP-INSP-A", "ROLE-004", "PERM-INSP-A"},
	{"RP-SUP-INSP-E", "ROLE-004", "PERM-INSP-E"},
	{"RP-SUP-ISS-R", "ROLE-004", "PERM-ISS-R"},
	{"RP-SUP-ISS-U", "ROLE-004", "PERM-ISS-U"},
	{"RP-SUP-LOG-R", "ROLE-004", "PERM-LOG-R"},
	{"RP-SUP-MSTR-R", "ROLE-004", "PERM-MSTR-R"},
	{"RP-SUP-PIC-R", "ROLE-004", "PERM-PIC-R"},

	// ── ROLE-005: Manager (baca semua, approve, export) ─────────────────────
	{"RP-MGR-INSP-R", "ROLE-005", "PERM-INSP-R"},
	{"RP-MGR-INSP-A", "ROLE-005", "PERM-INSP-A"},
	{"RP-MGR-INSP-E", "ROLE-005", "PERM-INSP-E"},
	{"RP-MGR-ISS-R", "ROLE-005", "PERM-ISS-R"},
	{"RP-MGR-ISS-U", "ROLE-005", "PERM-ISS-U"},
	{"RP-MGR-ISS-E", "ROLE-005", "PERM-ISS-E"},
	{"RP-MGR-LOG-R", "ROLE-005", "PERM-LOG-R"},
	{"RP-MGR-MSTR-R", "ROLE-005", "PERM-MSTR-R"},
	{"RP-MGR-PIC-R", "ROLE-005", "PERM-PIC-R"},
	{"RP-MGR-USR-R", "ROLE-005", "PERM-USR-R"},
}

func main() {
	cfg := config.Load()
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Ensure table exists
	_ = db.AutoMigrate(&authdomain.RolePermission{})

	// Clean up default issue read/create permissions for Auditor role
	_ = db.Where("\"RolePermissionID\" IN ?", []string{"RP-AUD-ISS-C", "RP-AUD-ISS-R"}).Delete(&authdomain.RolePermission{})

	inserted := 0
	skipped := 0

	for _, entry := range defaultRolePermissions {
		var count int64
		db.Model(&authdomain.RolePermission{}).
			Where("\"RolePermissionID\" = ?", entry.id).
			Count(&count)

		if count > 0 {
			skipped++
			continue
		}

		rp := authdomain.RolePermission{
			RolePermissionID:        entry.id,
			RoleID:                  entry.roleID,
			PermissionID:            entry.permissionID,
			IsAllowed:               true,
			RolePermissionCreatedAt: time.Now(),
			RolePermissionUpdatedAt: time.Now(),
			RolePermissionUpdatedBy: "USR-ADMIN-001",
		}

		if err := db.Create(&rp).Error; err != nil {
			log.Printf("  ❌ Failed to insert %s: %v", entry.id, err)
		} else {
			inserted++
			fmt.Printf("  ✔ Inserted: %s → %s [%s]\n", entry.roleID, entry.permissionID, entry.id)
		}
	}

	fmt.Printf("\n✅ Done! Inserted: %d, Skipped (already exist): %d\n", inserted, skipped)
}

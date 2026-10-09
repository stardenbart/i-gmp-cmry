package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/auth"
)

type fakeRolePerm struct{ allowed bool }

func (f fakeRolePerm) GetByRoleID(string) ([]auth.RolePermission, error)          { return nil, nil }
func (f fakeRolePerm) SetPermission(string, *auth.BulkSetPermissionRequest) error { return nil }
func (f fakeRolePerm) CheckPermission(string, string, string) (bool, error)       { return f.allowed, nil }

type fakeUserPerm struct{ override *bool }

func (f fakeUserPerm) GetByUserID(string) ([]auth.UserPermission, error) { return nil, nil }
func (f fakeUserPerm) SetPermissionOverrides(string, *auth.BulkSetUserPermissionRequest) error {
	return nil
}
func (f fakeUserPerm) CheckOverride(string, string, string) (*bool, error) { return f.override, nil }
func (f fakeUserPerm) ClearOverrides(string) error                         { return nil }

func boolp(b bool) *bool { return &b }

func runHasPermission(t *testing.T, roleID string, superAdmin bool, rp auth.RolePermissionUseCase, up auth.UserPermissionUseCase) bool {
	t.Helper()
	var got bool
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals(ContextKeyUserID, "USR-1")
		c.Locals(ContextKeyRoleID, roleID)
		c.Locals("isSuperAdmin", superAdmin)
		got = HasPermission(c, rp, up, "MOD-USR", "UPDATE")
		return nil
	})
	if _, err := app.Test(httptest.NewRequest("GET", "/", nil)); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestHasPermission(t *testing.T) {
	cases := []struct {
		name       string
		roleID     string
		superAdmin bool
		rp         fakeRolePerm
		up         fakeUserPerm
		want       bool
	}{
		{"super admin", "ROLE-000", true, fakeRolePerm{}, fakeUserPerm{}, true},
		{"role grants it", "ROLE-001", false, fakeRolePerm{allowed: true}, fakeUserPerm{}, true},
		{"role lacks it", "ROLE-003", false, fakeRolePerm{allowed: false}, fakeUserPerm{}, false},
		{"user override denies", "ROLE-001", false, fakeRolePerm{allowed: true}, fakeUserPerm{override: boolp(false)}, false},
		{"user override grants", "ROLE-003", false, fakeRolePerm{allowed: false}, fakeUserPerm{override: boolp(true)}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := runHasPermission(t, tc.roleID, tc.superAdmin, tc.rp, tc.up); got != tc.want {
				t.Fatalf("HasPermission = %v, want %v", got, tc.want)
			}
		})
	}
}

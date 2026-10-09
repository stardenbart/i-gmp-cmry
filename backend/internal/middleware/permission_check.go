package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/auth"
)

// ContextKeyCanManageUsers is set on the self-edit path of PUT /users/:id:
// true when the caller also holds MOD-USR UPDATE (Admin / Super Admin), so
// the handler may accept their own PIC kawasan changes.
const ContextKeyCanManageUsers = "can_manage_users"

// HasPermission answers the same question as PermissionMiddleware (Super
// Admin bypass, then user-level override, then role permission) without
// writing a response, for routes that only need to know.
func HasPermission(c *fiber.Ctx, rp auth.RolePermissionUseCase, up auth.UserPermissionUseCase, moduleID, permissionCode string) bool {
	roleID := GetRoleID(c)
	userID := GetUserID(c)

	if isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool); isSuperAdmin || roleID == "ROLE-000" || roleID == "SUPERADMIN" {
		return true
	}
	if up != nil && userID != "" {
		if override, err := up.CheckOverride(userID, moduleID, permissionCode); err == nil && override != nil {
			return *override
		}
	}
	if rp != nil && roleID != "" {
		allowed, err := rp.CheckPermission(roleID, moduleID, permissionCode)
		return err == nil && allowed
	}
	return false
}

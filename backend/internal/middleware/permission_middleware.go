package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/response"
)

// PermissionMiddleware dynamically checks whether the authenticated user or role has a specific
// permission (moduleID + permissionCode) based strictly on database records.
func PermissionMiddleware(
	rpUseCase auth.RolePermissionUseCase,
	upUseCase auth.UserPermissionUseCase,
	moduleID string,
	permissionCode string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := GetRoleID(c)
		userID := GetUserID(c)

		if roleID == "" && userID == "" {
			return response.Unauthorized(c, "unauthenticated: session identity not found in context")
		}

		// 0. SuperAdmin Bypass (ROLE-000 or isSuperAdmin context)
		isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
		if isSuperAdmin || roleID == "ROLE-000" || roleID == "SUPERADMIN" {
			return c.Next()
		}

		// 1. Check User-Level Permission Overrides First (Dynamic DB Query)
		if upUseCase != nil && userID != "" {
			override, err := upUseCase.CheckOverride(userID, moduleID, permissionCode)
			if err != nil {
				return response.InternalServerError(c, "user permission check failed", err.Error())
			}
			if override != nil {
				if !*override {
					// Fallback: if we are checking READ, see if they have UPDATE override
					if permissionCode == "READ" {
						updOverride, _ := upUseCase.CheckOverride(userID, moduleID, "UPDATE")
						if updOverride != nil && *updOverride {
							return c.Next()
						}
					}
					return response.Forbidden(c, "you do not have permission to perform this action (user restricted)")
				}
				return c.Next()
			}

			// Check if they have an override for UPDATE
			if permissionCode == "READ" {
				updOverride, _ := upUseCase.CheckOverride(userID, moduleID, "UPDATE")
				if updOverride != nil && *updOverride {
					return c.Next()
				}
			}
		}

		// 2. Check Role-Level Permissions (Dynamic DB Query)
		if rpUseCase != nil && roleID != "" {
			allowed, err := rpUseCase.CheckPermission(roleID, moduleID, permissionCode)
			if err != nil {
				return response.InternalServerError(c, "permission check failed", err.Error())
			}

			if allowed {
				return c.Next()
			}

			// Fallback: if we are checking READ, see if they have UPDATE role permission
			if permissionCode == "READ" {
				allowedUpdate, err := rpUseCase.CheckPermission(roleID, moduleID, "UPDATE")
				if err == nil && allowedUpdate {
					return c.Next()
				}

				// Fallback for Master Data READ (MOD-MSTR READ): allow users who have Inspection or Issue permissions to read Master dropdowns
				if moduleID == "MOD-MSTR" {
					allowedInsp, _ := rpUseCase.CheckPermission(roleID, "MOD-INSP", "READ")
					allowedInspC, _ := rpUseCase.CheckPermission(roleID, "MOD-INSP", "CREATE")
					allowedIss, _ := rpUseCase.CheckPermission(roleID, "MOD-ISS", "READ")
					allowedWowr, _ := rpUseCase.CheckPermission(roleID, "MOD-WOWR", "READ")
					if allowedInsp || allowedInspC || allowedIss || allowedWowr {
						return c.Next()
					}
				}
			}

			return response.Forbidden(c, "you do not have permission to perform this action")
		}

		return response.Forbidden(c, "you do not have permission to perform this action")
	}
}

// RequireSuperAdminMiddleware restricts access exclusively to SuperAdmin (ROLE-000).
func RequireSuperAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := GetRoleID(c)
		isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)

		if isSuperAdmin || roleID == "ROLE-000" || roleID == "SUPERADMIN" {
			return c.Next()
		}

		return response.Forbidden(c, "Hanya Super Admin yang diizinkan mengelola data Plant Master")
	}
}

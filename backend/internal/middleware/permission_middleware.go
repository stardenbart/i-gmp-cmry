package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/response"
)

// PermissionMiddleware checks whether the authenticated role has a specific
// permission (moduleID + permissionCode) before allowing access to the route.
//
// Usage:
//
//	router.GET("/path", authMiddleware, PermissionMiddleware(rpUseCase, "MOD-001", "READ"), handler)
func PermissionMiddleware(
	rpUseCase auth.RolePermissionUseCase,
	moduleID string,
	permissionCode string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := GetRoleID(c)
		if roleID == "" {
			return response.Unauthorized(c, "unauthenticated: role not found in context")
		}

		allowed, err := rpUseCase.CheckPermission(roleID, moduleID, permissionCode)
		if err != nil {
			response.InternalServerError(c, "permission check failed", err.Error())
			}

		if !allowed {
			return response.Forbidden(c, "you do not have permission to perform this action")
		}

		return c.Next()
	}
}

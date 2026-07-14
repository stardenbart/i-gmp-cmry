package middleware

import (
	"github.com/gin-gonic/gin"
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
) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := GetRoleID(c)
		if roleID == "" {
			response.Unauthorized(c, "unauthenticated: role not found in context")
			c.Abort()
			return
		}

		allowed, err := rpUseCase.CheckPermission(roleID, moduleID, permissionCode)
		if err != nil {
			response.InternalServerError(c, "permission check failed", err.Error())
			c.Abort()
			return
		}

		if !allowed {
			response.Forbidden(c, "you do not have permission to perform this action")
			c.Abort()
			return
		}

		c.Next()
	}
}

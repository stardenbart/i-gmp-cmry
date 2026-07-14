package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/response"
)

const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRoleID   = "role_id"
)

// AuthMiddleware validates the Bearer JWT token from the Authorization header.
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "invalid authorization header format, expected: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired token: "+err.Error())
			c.Abort()
			return
		}

		// Store claims in context for downstream handlers
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRoleID, claims.RoleID)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user's ID from gin context.
func GetUserID(c *gin.Context) string {
	return c.GetString(ContextKeyUserID)
}

// GetRoleID retrieves the authenticated user's role ID from gin context.
func GetRoleID(c *gin.Context) string {
	return c.GetString(ContextKeyRoleID)
}

package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/response"
)

const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRoleID   = "role_id"

	// CookieAccessToken/CookieRefreshToken/CookieCSRFToken are the httpOnly
	// (CookieCSRFToken excepted — see CSRFMiddleware) session cookies set by
	// AuthHandler.Login/Refresh and cleared by AuthHandler.Logout.
	CookieAccessToken  = "access_token"
	CookieRefreshToken = "refresh_token"
	CookieCSRFToken    = "csrf_token"
)

// AuthMiddleware validates the access token, read from (in order) the
// Authorization: Bearer header — kept for non-browser API clients that can't
// rely on cookies — or the access_token httpOnly cookie set at login. A
// token in the URL query string is deliberately NOT accepted here: it leaks
// into browser history, proxy/access logs and the Referer header, which a
// cookie or header does not.
func AuthMiddleware(jwtManager *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := ""
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr = parts[1]
			} else {
				return response.Unauthorized(c, "invalid authorization header format, expected: Bearer <token>")
			}
		}
		if tokenStr == "" {
			tokenStr = c.Cookies(CookieAccessToken)
		}

		if tokenStr == "" {
			return response.Unauthorized(c, "authentication required")
		}

		claims, err := jwtManager.Parse(tokenStr)
		if err != nil {
			return response.Unauthorized(c, "invalid or expired token: "+err.Error())
		}

		// Store claims in context for downstream handlers
		c.Locals(ContextKeyUserID, claims.UserID)
		c.Locals("userID", claims.UserID)
		c.Locals(ContextKeyUsername, claims.Username)
		c.Locals(ContextKeyRoleID, claims.RoleID)
		c.Locals("roleID", claims.RoleID)
		return c.Next()
	}
}

// GetUserID retrieves the authenticated user's ID from gin context.
func GetUserID(c *fiber.Ctx) string {
	val := c.Locals(ContextKeyUserID)
	if val == nil {
		return ""
	}
	return val.(string)
}

// GetRoleID retrieves the authenticated user's role ID from gin context.
func GetRoleID(c *fiber.Ctx) string {
	val := c.Locals(ContextKeyRoleID)
	if val == nil {
		return ""
	}
	return val.(string)
}

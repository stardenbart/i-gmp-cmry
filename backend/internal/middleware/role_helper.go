package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// IsAuditorRole returns true when roleID belongs to SuperAdmin, Admin, or Auditor.
func IsAuditorRole(roleID string) bool {
	r := strings.ToUpper(strings.TrimSpace(roleID))
	return r == "ROLE-000" || r == "ROLE-001" || r == "ROLE-002"
}

// RequireAdmin middleware checks if the authenticated user has SuperAdmin or Admin role.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := GetRoleID(c)
		r := strings.ToUpper(strings.TrimSpace(roleID))
		if r == "ROLE-000" || r == "ROLE-001" {
			return c.Next()
		}
		return c.Status(403).JSON(fiber.Map{
			"success": false,
			"message": "admin privilege required",
		})
	}
}

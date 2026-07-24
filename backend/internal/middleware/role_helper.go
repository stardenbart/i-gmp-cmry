package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// IsAuditorRole returns true when roleID belongs to an Auditor or Admin role.
// It is exported so that filter handlers (issuehandler, followup handler, etc.)
// can reuse the same logic without duplicating it.
func IsAuditorRole(roleID string) bool {
	r := strings.ToUpper(roleID)
	return r == "ROLE-001" || r == "ADM" || r == "ADMIN" || r == "1" ||
		r == "ROLE-002" || r == "AUDITOR" || r == "2"
}

// RequireAdmin middleware checks if the authenticated user has Admin role (ROLE-001 / ADMIN).
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := GetRoleID(c)
		r := strings.ToUpper(roleID)
		if r == "ROLE-001" || r == "ADM" || r == "ADMIN" || r == "1" {
			return c.Next()
		}
		return c.Status(403).JSON(fiber.Map{
			"success": false,
			"message": "admin privilege required",
		})
	}
}

package middleware

import (
	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
)

// PlantScopeMiddleware injects the authenticated user's PlantID scope into c.Locals("userPlantID").
// - SuperAdmin or unassigned global user: userPlantID = "" (bypass filter, can view all plants).
// - Admin / Auditor / Auditee / Supervisor / Manager: userPlantID = user.PlantID.
// - If user attempts to query a explicit plant_id different from their assigned PlantID -> 403 Forbidden.
func PlantScopeMiddleware(userRepo authdomain.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Next()
		}

		user, err := userRepo.FindByID(userID)
		if err != nil || user == nil {
			return c.Next()
		}

		roleID := user.RoleID
		isSuperAdmin := roleID == "ROLE-000" || roleID == "SUPERADMIN" || (user.Role != nil && user.Role.RoleName == "Super Admin")

		var assignedPlantID string
		if user.PlantID != nil {
			assignedPlantID = *user.PlantID
		}

		// If user is SuperAdmin or has no specific plant constraint, allow global query
		if isSuperAdmin {
			c.Locals("isSuperAdmin", true)
			c.Locals("userPlantID", "")
			return c.Next()
		}

		// Enforce single plant scope for non-SuperAdmin
		c.Locals("isSuperAdmin", false)
		c.Locals("userPlantID", assignedPlantID)

		// Check if explicit query param plant_id is sent and conflicts with assignedPlantID
		queryPlantID := c.Query("plant_id")
		if queryPlantID != "" && assignedPlantID != "" && queryPlantID != assignedPlantID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Anda tidak memiliki akses ke Plant ini",
			})
		}

		return c.Next()
	}
}

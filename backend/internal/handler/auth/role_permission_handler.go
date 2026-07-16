package auth

import (
	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
)

type RolePermissionHandler struct {
	uc authdomain.RolePermissionUseCase
}

func NewRolePermissionHandler(uc authdomain.RolePermissionUseCase) *RolePermissionHandler {
	return &RolePermissionHandler{uc: uc}
}

// @Summary      Get Role Permissions
// @Description  Get all permissions assigned to a role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /roles/{id}/permissions [get]
func (h *RolePermissionHandler) GetByRoleID(c *fiber.Ctx) error {
	roleID := c.Params("id")
	rps, err := h.uc.GetByRoleID(roleID)
	if err != nil {
		return response.BadRequest(c, "failed to get role permissions", err.Error())
	}
	return response.OK(c, "success", rps)
}

// @Summary      Bulk Set Role Permissions
// @Description  Set or update permissions for a role in bulk
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Param        body body interface{} true "Bulk Set Request"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /roles/{id}/permissions [put]
func (h *RolePermissionHandler) SetPermissions(c *fiber.Ctx) error {
	roleID := c.Params("id")
	actorID := middleware.GetUserID(c)

	var req authdomain.BulkSetPermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}
	req.RoleID = roleID

	if err := h.uc.SetPermission(actorID, &req); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	return response.OK(c, "permissions updated successfully", nil)
}

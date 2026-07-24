package auth

import (
	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
)

type UserPermissionHandler struct {
	uc authdomain.UserPermissionUseCase
}

func NewUserPermissionHandler(uc authdomain.UserPermissionUseCase) *UserPermissionHandler {
	return &UserPermissionHandler{uc: uc}
}

// @Summary      Get User Permission Overrides
// @Description  Get all permission overrides assigned to a user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users/{id}/permissions [get]
func (h *UserPermissionHandler) GetByUserID(c *fiber.Ctx) error {
	userID := c.Params("id")
	ups, err := h.uc.GetByUserID(userID)
	if err != nil {
		return response.BadRequest(c, "failed to get user permissions", err.Error())
	}
	return response.OK(c, "success", ups)
}

// @Summary      Set User Permission Overrides
// @Description  Set or update permission overrides for a user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        body body interface{} true "Bulk Set Request"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users/{id}/permissions [put]
func (h *UserPermissionHandler) SetPermissions(c *fiber.Ctx) error {
	userID := c.Params("id")
	actorID := middleware.GetUserID(c)

	var req authdomain.BulkSetUserPermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}
	req.UserID = userID

	if err := h.uc.SetPermissionOverrides(actorID, &req); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	return response.OK(c, "user permissions updated successfully", nil)
}

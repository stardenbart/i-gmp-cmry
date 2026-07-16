package auth

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/response"
)

type RoleHandler struct {
	roleUC authdomain.RoleUseCase
}

func NewRoleHandler(roleUC authdomain.RoleUseCase) *RoleHandler {
	return &RoleHandler{roleUC: roleUC}
}

// GetAll godoc
// @Summary      Get all roles
// @Description  Get paginated list of roles
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number"
// @Param        limit query int false "Limit per page"
// @Param        search query string false "Search term"
// @Success      200 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /roles [get]
func (h *RoleHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	roles, total, err := h.roleUC.GetAll(page, limit, search)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve roles", err)
	}

	return response.Paginated(c, "Roles retrieved successfully", roles, total, page, limit)
}

// GetByID godoc
// @Summary      Get role by ID
// @Description  Get a role by its ID
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /roles/{id} [get]
func (h *RoleHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	role, err := h.roleUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, "Role not found")
	}
	return response.OK(c, "Role retrieved successfully", role)
}

// Create godoc
// @Summary      Create role
// @Description  Create a new role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        body body interface{} true "Role data"
// @Success      201 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /roles [post]
func (h *RoleHandler) Create(c *fiber.Ctx) error {
	var role authdomain.Role
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, "Invalid request payload", err)
	}

	if err := h.roleUC.Create(&role); err != nil {
		return response.InternalServerError(c, "Failed to create role", err)
	}

	return response.Created(c, "Role created successfully", role)
}

// Update godoc
// @Summary      Update role
// @Description  Update an existing role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Param        body body interface{} true "Role data"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /roles/{id} [put]
func (h *RoleHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var role authdomain.Role
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, "Invalid request payload", err)
	}
	role.RoleID = id

	if err := h.roleUC.Update(&role); err != nil {
		return response.InternalServerError(c, "Failed to update role", err)
	}

	return response.OK(c, "Role updated successfully", role)
}

// Delete godoc
// @Summary      Delete role
// @Description  Delete a role by ID
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /roles/{id} [delete]
func (h *RoleHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.roleUC.Delete(id); err != nil {
		return response.InternalServerError(c, "Failed to delete role", err)
	}

	return response.OK(c, "Role deleted successfully", nil)
}

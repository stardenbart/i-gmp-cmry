package auth

import (
	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/response"
)

// UserFilterHandler handles GET /api/v1/users/filter.
type UserFilterHandler struct {
	uc authdomain.UserFilterUseCase
}

// NewUserFilterHandler creates a new UserFilterHandler.
func NewUserFilterHandler(uc authdomain.UserFilterUseCase) *UserFilterHandler {
	return &UserFilterHandler{uc: uc}
}

// GetFiltered handles the filter + facet request for users.
// @Summary      Filter users with facets
// @Description  Returns paginated users with facet counts. Requires MOD-USR READ permission.
// @Tags         users
// @Produce      json
// @Param        q               query  string  false  "Free-text search on Username/FullName/Email"
// @Param        role_id         query  string  false  "Filter by RoleID"
// @Param        role_id__in     query  string  false  "Comma-separated role IDs (OR)"
// @Param        department_id   query  string  false  "Filter by DepartmentID"
// @Param        user_status     query  string  false  "Active|Inactive|Suspended"
// @Param        date_from       query  string  false  "Created from"
// @Param        date_to         query  string  false  "Created to"
// @Param        sort_by         query  string  false  "created_at|username|full_name"
// @Param        sort_order      query  string  false  "asc or desc"
// @Param        page            query  int     false  "Page number"
// @Param        limit           query  int     false  "Items per page"
// @Success      200  {object}  response.APIResponse
// @Failure      500  {object}  response.APIResponse
// @Router       /api/v1/users/filter [get]
// @Security     BearerAuth
func (h *UserFilterHandler) GetFiltered(c *fiber.Ctx) error {
	f := authdomain.NewUserFilter(c)

	result, err := h.uc.GetFiltered(f)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch users", err.Error())
	}

	return response.OK(c, "success", result)
}

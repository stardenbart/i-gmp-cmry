package issuehandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
)

// FollowupFilterHandler handles GET /api/v1/issues/followup/filter.
type FollowupFilterHandler struct {
	uc issue.FollowupFilterUseCase
}

// NewFollowupFilterHandler creates a new FollowupFilterHandler.
func NewFollowupFilterHandler(uc issue.FollowupFilterUseCase) *FollowupFilterHandler {
	return &FollowupFilterHandler{uc: uc}
}

// GetFiltered handles the filter + facet request for issue followups.
// @Summary      Filter issue followups with facets
// @Description  Returns paginated issues pending followup with facets. Excludes Verified/Closed by default.
// @Tags         issues
// @Produce      json
// @Param        q               query  string  false  "Free-text search on Label"
// @Param        status          query  string  false  "Exact status"
// @Param        status__in      query  string  false  "Comma-separated statuses (OR)"
// @Param        wowr_status     query  string  false  "Filter by WOWR status"
// @Param        include_closed  query  string  false  "true to include Closed/Verified"
// @Param        date_from       query  string  false  "Created from"
// @Param        date_to         query  string  false  "Created to"
// @Param        due_from        query  string  false  "DueDate from"
// @Param        due_to          query  string  false  "DueDate to"
// @Param        sort_by         query  string  false  "created_at|due_date|status"
// @Param        sort_order      query  string  false  "asc or desc"
// @Param        page            query  int     false  "Page number"
// @Param        limit           query  int     false  "Items per page"
// @Success      200  {object}  response.APIResponse
// @Failure      500  {object}  response.APIResponse
// @Router       /api/v1/issues/followup/filter [get]
// @Security     BearerAuth
func (h *FollowupFilterHandler) GetFiltered(c *fiber.Ctx) error {
	f := issue.NewFollowupFilter(c)

	// Apply role-based scope: auditee only sees their own (PIC or delegated)
	actorRole := middleware.GetRoleID(c)
	actorID := middleware.GetUserID(c)
	if !middleware.IsAuditorRole(actorRole) {
		f.ScopeUserID = actorID
	}

	// Apply plant scope: non-SuperAdmin users can only see issues from their own plant
	userPlantID, _ := c.Locals("userPlantID").(string)
	if userPlantID != "" {
		f.PlantID = userPlantID
	}

	result, err := h.uc.GetFiltered(f)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch followup issues", err.Error())
	}

	return response.OK(c, "success", result)
}

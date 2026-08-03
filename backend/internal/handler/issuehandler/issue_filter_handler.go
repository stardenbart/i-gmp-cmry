package issuehandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
)

// IssueFilterHandler handles GET /api/v1/issues/filter.
type IssueFilterHandler struct {
	uc issue.IssueFilterUseCase
}

// NewIssueFilterHandler creates a new IssueFilterHandler.
func NewIssueFilterHandler(uc issue.IssueFilterUseCase) *IssueFilterHandler {
	return &IssueFilterHandler{uc: uc}
}

// GetFiltered handles the filter + facet request for issues.
// @Summary      Filter issues with facets
// @Description  Returns paginated issues with facet counts. Auditee can only see their own.
// @Tags         issues
// @Produce      json
// @Param        q                   query  string  false  "Free-text search on Label"
// @Param        status              query  string  false  "Exact status"
// @Param        status__in          query  string  false  "Comma-separated statuses (OR)"
// @Param        issue_pic_user_id   query  string  false  "Filter by PIC user"
// @Param        wowr_status         query  string  false  "Filter by WOWR status"
// @Param        needs_wo_wr         query  string  false  "true/false"
// @Param        date_from           query  string  false  "Created from"
// @Param        date_to             query  string  false  "Created to"
// @Param        due_from            query  string  false  "DueDate from"
// @Param        due_to              query  string  false  "DueDate to"
// @Param        sort_by             query  string  false  "created_at|due_date|status"
// @Param        sort_order          query  string  false  "asc or desc"
// @Param        page                query  int     false  "Page number"
// @Param        limit               query  int     false  "Items per page"
// @Success      200  {object}  response.APIResponse
// @Failure      500  {object}  response.APIResponse
// @Router       /api/v1/issues/filter [get]
// @Security     BearerAuth
func (h *IssueFilterHandler) GetFiltered(c *fiber.Ctx) error {
	f := issue.NewIssueFilter(c)

	// Apply role-based scope: auditee can only see their own issues
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
		return response.InternalServerError(c, "failed to fetch issues", err.Error())
	}

	return response.OK(c, "success", result)
}

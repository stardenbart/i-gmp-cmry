package inspectionhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
)

// InspectionFilterHandler handles GET /api/v1/inspections/filter.
type InspectionFilterHandler struct {
	uc inspection.InspectionFilterUseCase
}

// NewInspectionFilterHandler creates a new InspectionFilterHandler.
func NewInspectionFilterHandler(uc inspection.InspectionFilterUseCase) *InspectionFilterHandler {
	return &InspectionFilterHandler{uc: uc}
}

// GetFiltered handles the filter + facet request for inspections.
// @Summary      Filter inspections with facets
// @Description  Returns paginated inspections with facet counts. Auditee can only see their own.
// @Tags         inspections
// @Produce      json
// @Param        q               query  string  false  "Free-text search on SessionID"
// @Param        status          query  string  false  "Exact status filter"
// @Param        status__in      query  string  false  "Comma-separated statuses (OR)"
// @Param        area_id         query  string  false  "Filter by AreaID"
// @Param        kawasan_id      query  string  false  "Filter by KawasanID"
// @Param        inspector_id    query  string  false  "Filter by InspectorID"
// @Param        date_from       query  string  false  "Created from (RFC3339 or YYYY-MM-DD)"
// @Param        date_to         query  string  false  "Created to (RFC3339 or YYYY-MM-DD)"
// @Param        sort_by         query  string  false  "Sort field: created_at|updated_at|status"
// @Param        sort_order      query  string  false  "asc or desc"
// @Param        page            query  int     false  "Page number"
// @Param        limit           query  int     false  "Items per page (max 100)"
// @Success      200  {object}  response.APIResponse
// @Failure      500  {object}  response.APIResponse
// @Router       /api/v1/inspections/filter [get]
// @Security     BearerAuth
func (h *InspectionFilterHandler) GetFiltered(c *fiber.Ctx) error {
	f := inspection.NewInspectionFilter(c)

	// Apply role-based scope: auditee can only see their own inspections
	actorRole := middleware.GetRoleID(c)
	actorID := middleware.GetUserID(c)
	if !middleware.IsAuditorRole(actorRole) {
		f.ScopeInspectorID = actorID
	}

	// Apply plant scope: non-SuperAdmin users can only see inspections from their own plant
	userPlantID, _ := c.Locals("userPlantID").(string)
	if userPlantID != "" {
		f.PlantID = userPlantID
	}

	result, err := h.uc.GetFiltered(f)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch inspections", err.Error())
	}

	return response.OK(c, "success", result)
}

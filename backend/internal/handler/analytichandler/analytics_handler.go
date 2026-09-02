// Package analytichandler exposes the Custom KPI Visualization Builder's
// semantic-layer endpoints. Handlers here do HTTP parsing only — semantic
// validation, authorization, and query execution all live in
// analyticsusecase.QueryService (see the approved plan's layering rule).
package analytichandler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/analyticsusecase"
)

type AnalyticsHandler struct {
	service *analyticsusecase.QueryService
}

func NewAnalyticsHandler(service *analyticsusecase.QueryService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

// GetCatalog handles GET /api/v1/analytics/catalog — a read-only listing of
// the measures/dimensions a user is allowed to combine. No auth-scoping
// needed beyond "is logged in": the catalog itself carries no data, only
// field metadata (see QueryService.Catalog's doc comment on what it never
// includes).
func (h *AnalyticsHandler) GetCatalog(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "success",
		"data":    h.service.Catalog(),
	})
}

// analyticsQueryRequest is the wire shape for POST /api/v1/analytics/query.
type analyticsQueryRequest struct {
	Measures   []string                    `json:"measures"`
	Dimension  string                      `json:"dimension"`
	Dimension2 string                      `json:"dimension2"`
	Filters    []analyticsQueryFilterInput `json:"filters"`
	AreaID     string                      `json:"area_id"`
	StartDate  string                      `json:"start_date"`
	EndDate    string                      `json:"end_date"`
}

// analyticsQueryFilterInput is one ad-hoc "dimension = value" drill-down
// filter — see analyticsusecase.DimensionFilter.
type analyticsQueryFilterInput struct {
	DimensionID string `json:"dimension_id"`
	Value       string `json:"value"`
}

// RunQuery handles POST /api/v1/analytics/query. Body, not GET+query
// string, because the payload is expected to grow (Fase 2 calculations)
// without needing a new endpoint.
func (h *AnalyticsHandler) RunQuery(c *fiber.Ctx) error {
	var body analyticsQueryRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Request body tidak valid",
		})
	}

	filters := make([]analyticsusecase.DimensionFilter, 0, len(body.Filters))
	for _, f := range body.Filters {
		filters = append(filters, analyticsusecase.DimensionFilter{
			DimensionID: strings.TrimSpace(f.DimensionID),
			Value:       f.Value,
		})
	}

	scope := resolveScope(c)
	req := analyticsusecase.QueryRequest{
		Measures:   body.Measures,
		Dimension:  strings.TrimSpace(body.Dimension),
		Dimension2: strings.TrimSpace(body.Dimension2),
		Filters:    filters,
		AreaID:     strings.TrimSpace(body.AreaID),
		StartDate:  strings.TrimSpace(body.StartDate),
		EndDate:    strings.TrimSpace(body.EndDate),
	}

	result, err := h.service.RunQuery(c.Context(), scope, req)
	if err != nil {
		if svcErr, ok := err.(*analyticsusecase.ServiceError); ok {
			return c.Status(svcErr.Status).JSON(fiber.Map{"message": svcErr.Msg})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Terjadi kesalahan saat memproses permintaan",
		})
	}

	return c.JSON(fiber.Map{
		"message": "success",
		"data":    result,
	})
}

// resolveScope reads the auth/plant-scope middleware locals (set by
// AuthMiddleware + PlantScopeMiddleware, already applied to this route
// group) into a plain analyticsusecase.Scope — this is the ONLY place
// *fiber.Ctx is touched for authorization, keeping the service testable
// without a fake Ctx.
func resolveScope(c *fiber.Ctx) analyticsusecase.Scope {
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	roleID := middleware.GetRoleID(c)
	if roleID == "ROLE-000" || roleID == "SUPERADMIN" {
		isSuperAdmin = true
	}
	userPlantID, _ := c.Locals("userPlantID").(string)
	return analyticsusecase.Scope{
		IsSuperAdmin: isSuperAdmin,
		UserPlantID:  userPlantID,
		RoleID:       roleID,
		UserID:       middleware.GetUserID(c),
		QueryPlantID: c.Query("plant_id"),
	}
}

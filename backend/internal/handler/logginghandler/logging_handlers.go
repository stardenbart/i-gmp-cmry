package logginghandler

import (
	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
)

// ── LoginLog Handler ──────────────────────────────────────────────────────

type LoginLogHandler struct{ uc logdomain.LoginLogUseCase }

func NewLoginLogHandler(uc logdomain.LoginLogUseCase) *LoginLogHandler {
	return &LoginLogHandler{uc: uc}
}

// @Summary Get all login logs
// @Description Get a paginated list of login logs, optionally filtered by user_id
// @Tags login-logs
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param user_id query string false "Filter by User ID"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/logs/login [get]
// @Security BearerAuth
func (h *LoginLogHandler) GetAll(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	plantID := c.Query("plant_id")
	userPlantID, _ := c.Locals("userPlantID").(string)
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	if !isSuperAdmin && userPlantID != "" {
		plantID = userPlantID
	}
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("user_id"), plantID)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch login logs", err.Error())
	}
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

// @Summary Get login log by ID
// @Description Get a specific login log by its ID
// @Tags login-logs
// @Accept json
// @Produce json
// @Param id path string true "Log ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/logs/login/{id} [get]
// @Security BearerAuth
func (h *LoginLogHandler) GetByID(c *fiber.Ctx) error {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil {
		return response.NotFound(c, "log not found")
	}
	return response.OK(c, "success", item)
}

// ── ActivityLog Handler ───────────────────────────────────────────────────

type ActivityLogHandler struct{ uc logdomain.ActivityLogUseCase }

func NewActivityLogHandler(uc logdomain.ActivityLogUseCase) *ActivityLogHandler {
	return &ActivityLogHandler{uc: uc}
}

// @Summary Get all activity logs
// @Description Get a paginated list of activity logs, optionally filtered by user_id, module_id, or action
// @Tags activity-logs
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param user_id query string false "Filter by User ID"
// @Param module_id query string false "Filter by Module ID"
// @Param action query string false "Filter by Action"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/logs/activity [get]
// @Security BearerAuth
func (h *ActivityLogHandler) GetAll(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	search := c.Query("search")
	if search == "" {
		search = c.Query("action")
	}
	plantID := c.Query("plant_id")
	userPlantID, _ := c.Locals("userPlantID").(string)
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	if !isSuperAdmin && userPlantID != "" {
		plantID = userPlantID
	}
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("user_id"), c.Query("module_id"), search, plantID)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch activity logs", err.Error())
	}
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

// @Summary Get activity log by ID
// @Description Get a specific activity log by its ID
// @Tags activity-logs
// @Accept json
// @Produce json
// @Param id path string true "Log ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/logs/activity/{id} [get]
// @Security BearerAuth
func (h *ActivityLogHandler) GetByID(c *fiber.Ctx) error {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil {
		return response.NotFound(c, "log not found")
	}
	return response.OK(c, "success", item)
}

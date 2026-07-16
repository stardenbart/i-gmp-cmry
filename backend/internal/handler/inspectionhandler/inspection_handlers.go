package inspectionhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

// ── Inspection Header Handler ─────────────────────────────────────────────

type InspectionHeaderHandler struct{ uc inspection.InspectionHeaderUseCase }

func NewInspectionHeaderHandler(uc inspection.InspectionHeaderUseCase) *InspectionHeaderHandler {
	return &InspectionHeaderHandler{uc: uc}
}

// @Summary Get all inspections
// @Description Get a paginated list of inspections with optional filters
// @Tags Inspections
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Param area_id query string false "Area ID filter"
// @Param status query string false "Status filter"
// @Param inspector_id query string false "Inspector ID filter"
// @Success 200 {object} response.APIResponse "success"
// @Failure 500 {object} response.APIResponse "failed to fetch inspections"
// @Router /inspections [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("status"), c.Query("inspector_id"))
	if err != nil { return response.InternalServerError(c, "failed to fetch inspections", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

// @Summary Get inspection by ID
// @Description Retrieve a specific inspection header by its ID
// @Tags Inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} response.APIResponse "success"
// @Failure 404 {object} response.APIResponse "inspection not found"
// @Router /inspections/{id} [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "inspection not found") }
	return response.OK(c, "success", item)
}

// @Summary Get area status
// @Description Calculate and get the inspection status progress for a specific area
// @Tags Inspections
// @Accept json
// @Produce json
// @Param areaId path string true "Area ID"
// @Success 200 {object} response.APIResponse "success"
// @Failure 500 {object} response.APIResponse "failed to calculate area status"
// @Router /inspections/area/{areaId}/status [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) GetAreaStatus(c *fiber.Ctx) error  {
	progress, err := h.uc.GetAreaStatus(c.Params("areaId"))
	if err != nil { return response.InternalServerError(c, "failed to calculate area status", err.Error()) }
	return response.OK(c, "success", progress)
}

// @Summary Create inspection
// @Description Create a new inspection header
// @Tags Inspections
// @Accept json
// @Produce json
// @Param body body inspection.CreateInspectionRequest true "Inspection creation request"
// @Success 201 {object} response.APIResponse "inspection created"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspections [post]
// @Security BearerAuth
func (h *InspectionHeaderHandler) Create(c *fiber.Ctx) error  {
	var req inspection.CreateInspectionRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if errs := validator.Validate(&req); errs != nil { return response.BadRequest(c, "validation failed", errs) }
	inspectorID := middleware.GetUserID(c)
	item, err := h.uc.Create(inspectorID, &req)
	if err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "inspection created", item)
}

// @Summary Update inspection status
// @Description Update the status of an existing inspection
// @Tags Inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Param body body inspection.UpdateInspectionStatusRequest true "Status update request"
// @Success 200 {object} response.APIResponse "status updated"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspections/{id}/status [patch]
// @Security BearerAuth
func (h *InspectionHeaderHandler) UpdateStatus(c *fiber.Ctx) error  {
	var req inspection.UpdateInspectionStatusRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if errs := validator.Validate(&req); errs != nil { return response.BadRequest(c, "validation failed", errs) }
	actorID := middleware.GetUserID(c)
	item, err := h.uc.UpdateStatus(c.Params("id"), actorID, &req)
	if err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "status updated", item)
}

// @Summary Delete inspection
// @Description Delete an inspection by ID
// @Tags Inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} response.APIResponse "inspection deleted"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspections/{id} [delete]
// @Security BearerAuth
func (h *InspectionHeaderHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "inspection deleted", nil)
}

// ── Inspection Result Handler ─────────────────────────────────────────────

type InspectionResultHandler struct{ uc inspection.InspectionResultUseCase }

func NewInspectionResultHandler(uc inspection.InspectionResultUseCase) *InspectionResultHandler {
	return &InspectionResultHandler{uc: uc}
}

// @Summary Get inspection results by inspection ID
// @Description Retrieve all results associated with a specific inspection
// @Tags Inspection Results
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} response.APIResponse "success"
// @Failure 500 {object} response.APIResponse "failed to fetch results"
// @Router /inspections/{id}/results [get]
// @Security BearerAuth
func (h *InspectionResultHandler) GetByInspectionID(c *fiber.Ctx) error  {
	items, err := h.uc.GetByInspectionID(c.Params("id"))
	if err != nil { return response.InternalServerError(c, "failed to fetch results", err.Error()) }
	return response.OK(c, "success", items)
}

// @Summary Bulk save inspection results
// @Description Save multiple results for a specific inspection in bulk
// @Tags Inspection Results
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Param body body inspection.BulkSaveResultRequest true "Bulk save request"
// @Success 200 {object} response.APIResponse "results saved"
// @Failure 400 {object} response.APIResponse "bad request"
// @Failure 500 {object} response.APIResponse "internal server error"
// @Router /inspections/{id}/results/bulk [post]
// @Security BearerAuth
func (h *InspectionResultHandler) BulkSave(c *fiber.Ctx) error  {
	var req inspection.BulkSaveResultRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	req.InspectionID = c.Params("id")
	if err := h.uc.BulkSave(&req); err != nil { return response.InternalServerError(c, err.Error(), nil) }
	return response.OK(c, "results saved", nil)
}

// @Summary Update inspection result
// @Description Update a specific inspection result
// @Tags Inspection Results
// @Accept json
// @Produce json
// @Param result_id path string true "Result ID"
// @Param body body inspection.SaveResultRequest true "Update result request"
// @Success 200 {object} response.APIResponse "result updated"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspection-results/{result_id} [put]
// @Security BearerAuth
func (h *InspectionResultHandler) Update(c *fiber.Ctx) error  {
	var req inspection.SaveResultRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item, err := h.uc.Update(c.Params("result_id"), &req)
	if err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "result updated", item)
}

// @Summary Delete inspection result
// @Description Delete a specific inspection result
// @Tags Inspection Results
// @Accept json
// @Produce json
// @Param result_id path string true "Result ID"
// @Success 200 {object} response.APIResponse "result deleted"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspection-results/{result_id} [delete]
// @Security BearerAuth
func (h *InspectionResultHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("result_id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "result deleted", nil)
}

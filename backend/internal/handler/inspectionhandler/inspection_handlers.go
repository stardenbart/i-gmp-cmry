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

func (h *InspectionHeaderHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("status"), c.Query("inspector_id"))
	if err != nil { return response.InternalServerError(c, "failed to fetch inspections", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *InspectionHeaderHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "inspection not found") }
	return response.OK(c, "success", item)
}

func (h *InspectionHeaderHandler) Create(c *fiber.Ctx) error  {
	var req inspection.CreateInspectionRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if errs := validator.Validate(&req); errs != nil { return response.BadRequest(c, "validation failed", errs) }
	inspectorID := middleware.GetUserID(c)
	item, err := h.uc.Create(inspectorID, &req)
	if err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "inspection created", item)
}

func (h *InspectionHeaderHandler) UpdateStatus(c *fiber.Ctx) error  {
	var req inspection.UpdateInspectionStatusRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if errs := validator.Validate(&req); errs != nil { return response.BadRequest(c, "validation failed", errs) }
	actorID := middleware.GetUserID(c)
	item, err := h.uc.UpdateStatus(c.Params("id"), actorID, &req)
	if err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "status updated", item)
}

func (h *InspectionHeaderHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "inspection deleted", nil)
}

// ── Inspection Result Handler ─────────────────────────────────────────────

type InspectionResultHandler struct{ uc inspection.InspectionResultUseCase }

func NewInspectionResultHandler(uc inspection.InspectionResultUseCase) *InspectionResultHandler {
	return &InspectionResultHandler{uc: uc}
}

func (h *InspectionResultHandler) GetByInspectionID(c *fiber.Ctx) error  {
	items, err := h.uc.GetByInspectionID(c.Params("id"))
	if err != nil { return response.InternalServerError(c, "failed to fetch results", err.Error()) }
	return response.OK(c, "success", items)
}

func (h *InspectionResultHandler) BulkSave(c *fiber.Ctx) error  {
	var req inspection.BulkSaveResultRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	req.InspectionID = c.Params("id")
	if err := h.uc.BulkSave(&req); err != nil { return response.InternalServerError(c, err.Error(), nil) }
	return response.OK(c, "results saved", nil)
}

func (h *InspectionResultHandler) Update(c *fiber.Ctx) error  {
	var req inspection.SaveResultRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item, err := h.uc.Update(c.Params("result_id"), &req)
	if err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "result updated", item)
}

func (h *InspectionResultHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("result_id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "result deleted", nil)
}

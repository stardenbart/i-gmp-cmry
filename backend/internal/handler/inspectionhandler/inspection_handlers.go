package inspectionhandler

import (
	"github.com/gin-gonic/gin"
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

func (h *InspectionHeaderHandler) GetAll(c *gin.Context) {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("status"), c.Query("inspector_id"))
	if err != nil { response.InternalServerError(c, "failed to fetch inspections", err.Error()); return }
	response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *InspectionHeaderHandler) GetByID(c *gin.Context) {
	item, err := h.uc.GetByID(c.Param("id"))
	if err != nil { response.NotFound(c, "inspection not found"); return }
	response.OK(c, "success", item)
}

func (h *InspectionHeaderHandler) Create(c *gin.Context) {
	var req inspection.CreateInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid body", err.Error()); return }
	if errs := validator.Validate(&req); errs != nil { response.BadRequest(c, "validation failed", errs); return }
	inspectorID := middleware.GetUserID(c)
	item, err := h.uc.Create(inspectorID, &req)
	if err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.Created(c, "inspection created", item)
}

func (h *InspectionHeaderHandler) UpdateStatus(c *gin.Context) {
	var req inspection.UpdateInspectionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid body", err.Error()); return }
	if errs := validator.Validate(&req); errs != nil { response.BadRequest(c, "validation failed", errs); return }
	actorID := middleware.GetUserID(c)
	item, err := h.uc.UpdateStatus(c.Param("id"), actorID, &req)
	if err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "status updated", item)
}

func (h *InspectionHeaderHandler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Param("id")); err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "inspection deleted", nil)
}

// ── Inspection Result Handler ─────────────────────────────────────────────

type InspectionResultHandler struct{ uc inspection.InspectionResultUseCase }

func NewInspectionResultHandler(uc inspection.InspectionResultUseCase) *InspectionResultHandler {
	return &InspectionResultHandler{uc: uc}
}

func (h *InspectionResultHandler) GetByInspectionID(c *gin.Context) {
	items, err := h.uc.GetByInspectionID(c.Param("id"))
	if err != nil { response.InternalServerError(c, "failed to fetch results", err.Error()); return }
	response.OK(c, "success", items)
}

func (h *InspectionResultHandler) BulkSave(c *gin.Context) {
	var req inspection.BulkSaveResultRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid body", err.Error()); return }
	req.InspectionID = c.Param("id")
	if err := h.uc.BulkSave(&req); err != nil { response.InternalServerError(c, err.Error(), nil); return }
	response.OK(c, "results saved", nil)
}

func (h *InspectionResultHandler) Update(c *gin.Context) {
	var req inspection.SaveResultRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid body", err.Error()); return }
	item, err := h.uc.Update(c.Param("result_id"), &req)
	if err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "result updated", item)
}

func (h *InspectionResultHandler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Param("result_id")); err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "result deleted", nil)
}

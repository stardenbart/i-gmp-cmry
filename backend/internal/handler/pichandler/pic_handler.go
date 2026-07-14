package pichandler

import (
	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

type PICMappingHandler struct{ uc pic.PICMappingUseCase }

func NewPICMappingHandler(uc pic.PICMappingUseCase) *PICMappingHandler {
	return &PICMappingHandler{uc: uc}
}

func (h *PICMappingHandler) GetAll(c *gin.Context) {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("kawasan_id"))
	if err != nil { response.InternalServerError(c, "failed to fetch PIC mappings", err.Error()); return }
	response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *PICMappingHandler) GetByID(c *gin.Context) {
	item, err := h.uc.GetByID(c.Param("id"))
	if err != nil { response.NotFound(c, "PIC mapping not found"); return }
	response.OK(c, "success", item)
}

func (h *PICMappingHandler) Create(c *gin.Context) {
	var req pic.CreatePICMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid body", err.Error()); return }
	if errs := validator.Validate(&req); errs != nil { response.BadRequest(c, "validation failed", errs); return }
	item, err := h.uc.Create(&req)
	if err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.Created(c, "PIC mapping created", item)
}

func (h *PICMappingHandler) Update(c *gin.Context) {
	var req pic.UpdatePICMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid body", err.Error()); return }
	item, err := h.uc.Update(c.Param("id"), &req)
	if err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "PIC mapping updated", item)
}

func (h *PICMappingHandler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Param("id")); err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "PIC mapping deleted", nil)
}

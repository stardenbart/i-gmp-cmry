package pichandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

type PICMappingHandler struct{ uc pic.PICMappingUseCase }

func NewPICMappingHandler(uc pic.PICMappingUseCase) *PICMappingHandler {
	return &PICMappingHandler{uc: uc}
}

func (h *PICMappingHandler) GetAll(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("kawasan_id"))
	if err != nil {
		return response.InternalServerError(c, "failed to fetch PIC mappings", err.Error())
	}
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *PICMappingHandler) GetByID(c *fiber.Ctx) error {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil {
		return response.NotFound(c, "PIC mapping not found")
	}
	return response.OK(c, "success", item)
}

func (h *PICMappingHandler) Create(c *fiber.Ctx) error {
	var req pic.CreatePICMappingRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	item, err := h.uc.Create(&req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.Created(c, "PIC mapping created", item)
}

func (h *PICMappingHandler) Update(c *fiber.Ctx) error {
	var req pic.UpdatePICMappingRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	item, err := h.uc.Update(c.Params("id"), &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "PIC mapping updated", item)
}

func (h *PICMappingHandler) Delete(c *fiber.Ctx) error {
	if err := h.uc.Delete(c.Params("id")); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "PIC mapping deleted", nil)
}

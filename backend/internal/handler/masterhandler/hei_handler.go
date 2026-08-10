package masterhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
)

type HEIHandler struct {
	uc master.HEIUseCase
}

func NewHEIHandler(uc master.HEIUseCase) *HEIHandler {
	return &HEIHandler{uc: uc}
}

func (h *HEIHandler) GetAll(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	category := c.Query("category")
	search := c.Query("search")

	items, total, err := h.uc.GetAll(p.Page, p.Limit, category, search)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch HEI master items", err.Error())
	}
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *HEIHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.uc.GetCategories(c.Context())
	if err != nil {
		return response.InternalServerError(c, "failed to fetch HEI categories", err.Error())
	}
	return response.OK(c, "success", categories)
}

func (h *HEIHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.uc.GetByID(id)
	if err != nil || item == nil {
		return response.NotFound(c, "item HEI tidak ditemukan")
	}
	return response.OK(c, "success", item)
}

func (h *HEIHandler) Create(c *fiber.Ctx) error {
	var req master.CreateHEIRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}

	item, err := h.uc.Create(&req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.Created(c, "item HEI berhasil dibuat", item)
}

func (h *HEIHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req master.UpdateHEIRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}

	item, err := h.uc.Update(id, &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "item HEI berhasil diperbarui", item)
}

func (h *HEIHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.Delete(id); err != nil {
		return response.InternalServerError(c, "failed to delete item HEI", err.Error())
	}
	return response.OK(c, "item HEI berhasil dihapus", nil)
}

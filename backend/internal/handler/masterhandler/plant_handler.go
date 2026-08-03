package masterhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
)

type PlantHandler struct {
	uc master.PlantUseCase
}

func NewPlantHandler(uc master.PlantUseCase) *PlantHandler {
	return &PlantHandler{uc: uc}
}

func (h *PlantHandler) GetAll(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	search := c.Query("search", "")

	plants, total, err := h.uc.GetAll(p.Page, p.Limit, search)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch plants", err.Error())
	}

	return response.Paginated(c, "success", plants, total, p.Page, p.Limit)
}

func (h *PlantHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	plant, err := h.uc.GetByID(id)
	if err != nil {
		return response.NotFound(c, "Plant tidak ditemukan")
	}
	return response.OK(c, "success", plant)
}

func (h *PlantHandler) Create(c *fiber.Ctx) error {
	var req master.CreatePlantRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Payload tidak valid", err.Error())
	}

	plant, err := h.uc.Create(&req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	return response.Created(c, "Plant berhasil dibuat", plant)
}

func (h *PlantHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req master.UpdatePlantRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Payload tidak valid", err.Error())
	}

	plant, err := h.uc.Update(id, &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	return response.OK(c, "Plant berhasil diperbarui", plant)
}

func (h *PlantHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.Delete(id); err != nil {
		return response.InternalServerError(c, "failed to delete plant", err.Error())
	}
	return response.OK(c, "Plant berhasil dihapus", nil)
}

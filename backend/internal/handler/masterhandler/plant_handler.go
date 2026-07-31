package masterhandler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/master"
)

type PlantHandler struct {
	uc master.PlantUseCase
}

func NewPlantHandler(uc master.PlantUseCase) *PlantHandler {
	return &PlantHandler{uc: uc}
}

func (h *PlantHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")

	plants, total, err := h.uc.GetAll(page, limit, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    plants,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *PlantHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	plant, err := h.uc.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Plant tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": plant})
}

func (h *PlantHandler) Create(c *fiber.Ctx) error {
	var req master.CreatePlantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payload tidak valid"})
	}

	plant, err := h.uc.Create(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": plant})
}

func (h *PlantHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req master.UpdatePlantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payload tidak valid"})
	}

	plant, err := h.uc.Update(id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": plant})
}

func (h *PlantHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Plant berhasil dihapus"})
}

package issuehandler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/issue"
)

type HEIHandler struct {
	habitUC        issue.HabitUseCase
	equipmentUC    issue.EquipmentUseCase
	infraUC        issue.InfrastructureUseCase
}

func NewHEIHandler(hUC issue.HabitUseCase, eqUC issue.EquipmentUseCase, infUC issue.InfrastructureUseCase) *HEIHandler {
	return &HEIHandler{
		habitUC:     hUC,
		equipmentUC: eqUC,
		infraUC:     infUC,
	}
}

// ── Habit Handlers ─────────────────────────────────────────────────────────

func (h *HEIHandler) GetAllHabits(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")

	habits, total, err := h.habitUC.GetAll(page, limit, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    habits,
		"meta":    fiber.Map{"page": page, "limit": limit, "total": total},
	})
}

func (h *HEIHandler) GetHabitByID(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.habitUC.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Habit tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) CreateHabit(c *fiber.Ctx) error {
	var req issue.CreateHabitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payload tidak valid"})
	}
	item, err := h.habitUC.Create(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) UpdateHabit(c *fiber.Ctx) error {
	id := c.Params("id")
	var req issue.UpdateHabitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payload tidak valid"})
	}
	item, err := h.habitUC.Update(id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) DeleteHabit(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.habitUC.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Habit berhasil dihapus"})
}

// ── Equipment Handlers ─────────────────────────────────────────────────────

func (h *HEIHandler) GetAllEquipments(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	kawasanID := c.Query("kawasan_id", "")
	search := c.Query("search", "")

	equipments, total, err := h.equipmentUC.GetAll(page, limit, kawasanID, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    equipments,
		"meta":    fiber.Map{"page": page, "limit": limit, "total": total},
	})
}

func (h *HEIHandler) GetEquipmentByID(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.equipmentUC.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Equipment tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) CreateEquipment(c *fiber.Ctx) error {
	var req issue.CreateEquipmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payload tidak valid"})
	}
	item, err := h.equipmentUC.Create(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) UpdateEquipment(c *fiber.Ctx) error {
	id := c.Params("id")
	var req issue.UpdateEquipmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payload tidak valid"})
	}
	item, err := h.equipmentUC.Update(id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) DeleteEquipment(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.equipmentUC.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Equipment berhasil dihapus"})
}

// ── Infrastructure Handlers ────────────────────────────────────────────────

func (h *HEIHandler) GetAllInfrastructures(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	kawasanID := c.Query("kawasan_id", "")
	search := c.Query("search", "")

	infras, total, err := h.infraUC.GetAll(page, limit, kawasanID, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    infras,
		"meta":    fiber.Map{"page": page, "limit": limit, "total": total},
	})
}

func (h *HEIHandler) GetInfrastructureByID(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.infraUC.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Infrastructure tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) CreateInfrastructure(c *fiber.Ctx) error {
	var req issue.CreateInfrastructureRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payload tidak valid"})
	}
	item, err := h.infraUC.Create(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) UpdateInfrastructure(c *fiber.Ctx) error {
	id := c.Params("id")
	var req issue.UpdateInfrastructureRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payload tidak valid"})
	}
	item, err := h.infraUC.Update(id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": item})
}

func (h *HEIHandler) DeleteInfrastructure(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.infraUC.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Infrastructure berhasil dihapus"})
}

package masterhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

type SettingHandler struct{ uc master.SettingUseCase }

func NewSettingHandler(uc master.SettingUseCase) *SettingHandler {
	return &SettingHandler{uc: uc}
}

func getPlantID(c *fiber.Ctx) string {
	plantID := c.Query("plant_id")
	userPlantID, _ := c.Locals("userPlantID").(string)
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	if !isSuperAdmin && userPlantID != "" {
		plantID = userPlantID
	}
	return plantID
}

// GetAll returns all system settings (encrypted values are masked as ***)
func (h *SettingHandler) GetAll(c *fiber.Ctx) error {
	plantID := getPlantID(c)
	items, err := h.uc.GetAll(plantID)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch settings", err.Error())
	}
	return response.OK(c, "success", items)
}

// GetByKey returns a single setting by its key
func (h *SettingHandler) GetByKey(c *fiber.Ctx) error {
	plantID := getPlantID(c)
	item, err := h.uc.GetByKey(c.Params("key"), plantID)
	if err != nil {
		return response.NotFound(c, "setting not found")
	}
	return response.OK(c, "success", item)
}

// Update allows Admin to update a setting value dynamically
func (h *SettingHandler) Update(c *fiber.Ctx) error {
	var req master.UpdateSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}

	plantID := getPlantID(c)
	updatedBy := middleware.GetUserID(c)
	item, err := h.uc.Update(c.Params("key"), &req, updatedBy, plantID)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "setting updated", item)
}

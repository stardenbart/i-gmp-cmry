package auth

import (
	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/response"
)

type ModuleHandler struct {
	uc authdomain.ModuleUseCase
}

func NewModuleHandler(uc authdomain.ModuleUseCase) *ModuleHandler {
	return &ModuleHandler{uc: uc}
}

// @Summary      Get all modules
// @Description  Get all modules along with their permissions
// @Tags         Modules
// @Accept       json
// @Produce      json
// @Success      200 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /modules [get]
func (h *ModuleHandler) GetAll(c *fiber.Ctx) error {
	modules, err := h.uc.GetAll()
	if err != nil {
		return response.InternalServerError(c, "failed to fetch modules", err.Error())
	}
	return response.OK(c, "success", modules)
}

package masterhandler

import (
	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

type SettingHandler struct{ uc master.SettingUseCase }

func NewSettingHandler(uc master.SettingUseCase) *SettingHandler {
	return &SettingHandler{uc: uc}
}

// GetAll returns all system settings (encrypted values are masked as ***)
func (h *SettingHandler) GetAll(c *gin.Context) {
	items, err := h.uc.GetAll()
	if err != nil {
		response.InternalServerError(c, "failed to fetch settings", err.Error())
		return
	}
	response.OK(c, "success", items)
}

// GetByKey returns a single setting by its key
func (h *SettingHandler) GetByKey(c *gin.Context) {
	item, err := h.uc.GetByKey(c.Param("key"))
	if err != nil {
		response.NotFound(c, "setting not found")
		return
	}
	response.OK(c, "success", item)
}

// Update allows Admin to update a setting value dynamically
func (h *SettingHandler) Update(c *gin.Context) {
	var req master.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body", err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}

	updatedBy := middleware.GetUserID(c)
	item, err := h.uc.Update(c.Param("key"), &req, updatedBy)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.OK(c, "setting updated", item)
}

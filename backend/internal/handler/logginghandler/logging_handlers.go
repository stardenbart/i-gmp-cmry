package logginghandler

import (
	"github.com/gin-gonic/gin"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
)

// ── LoginLog Handler ──────────────────────────────────────────────────────

type LoginLogHandler struct{ uc logdomain.LoginLogUseCase }

func NewLoginLogHandler(uc logdomain.LoginLogUseCase) *LoginLogHandler { return &LoginLogHandler{uc: uc} }

func (h *LoginLogHandler) GetAll(c *gin.Context) {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("user_id"))
	if err != nil { response.InternalServerError(c, "failed to fetch login logs", err.Error()); return }
	response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *LoginLogHandler) GetByID(c *gin.Context) {
	item, err := h.uc.GetByID(c.Param("id"))
	if err != nil { response.NotFound(c, "log not found"); return }
	response.OK(c, "success", item)
}

// ── ActivityLog Handler ───────────────────────────────────────────────────

type ActivityLogHandler struct{ uc logdomain.ActivityLogUseCase }

func NewActivityLogHandler(uc logdomain.ActivityLogUseCase) *ActivityLogHandler {
	return &ActivityLogHandler{uc: uc}
}

func (h *ActivityLogHandler) GetAll(c *gin.Context) {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("user_id"), c.Query("module_id"), c.Query("action"))
	if err != nil { response.InternalServerError(c, "failed to fetch activity logs", err.Error()); return }
	response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *ActivityLogHandler) GetByID(c *gin.Context) {
	item, err := h.uc.GetByID(c.Param("id"))
	if err != nil { response.NotFound(c, "log not found"); return }
	response.OK(c, "success", item)
}

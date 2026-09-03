package kpisharehandler

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/kpishare"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/kpishareusecase"
	"github.com/monitoring-system/backend/pkg/response"
)

type Handler struct {
	service  *kpishareusecase.Service
	activity logdomain.ActivityLogUseCase
}

func New(service *kpishareusecase.Service, activity logdomain.ActivityLogUseCase) *Handler {
	return &Handler{service: service, activity: activity}
}

func actorFromContext(c *fiber.Ctx) kpishare.Actor {
	plantID, _ := c.Locals("userPlantID").(string)
	return kpishare.Actor{UserID: middleware.GetUserID(c), RoleID: middleware.GetRoleID(c), PlantID: plantID}
}

func writeError(c *fiber.Ctx, err error) error {
	status, message := kpishareusecase.ErrorStatus(err)
	if status == http.StatusNotFound {
		return response.NotFound(c, message)
	}
	if status == http.StatusForbidden {
		return response.Forbidden(c, message)
	}
	if status == http.StatusBadRequest {
		return response.BadRequest(c, message, nil)
	}
	return response.InternalServerError(c, message, nil)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req kpishare.CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Request body tidak valid", nil)
	}
	created, err := h.service.Create(actorFromContext(c), req)
	if err != nil {
		return writeError(c, err)
	}
	h.recordActivity(c, created.ShareID, "CREATE", "Membuat link publik Dashboard KPI untuk plant "+created.PlantID)
	return response.Created(c, "Link publik berhasil dibuat. Salin link sekarang karena token hanya ditampilkan sekali.", created)
}

func (h *Handler) List(c *fiber.Ctx) error {
	items, err := h.service.List(actorFromContext(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.OK(c, "success", items)
}

func (h *Handler) Revoke(c *fiber.Ctx) error {
	if err := h.service.Revoke(actorFromContext(c), c.Params("shareId")); err != nil {
		return writeError(c, err)
	}
	h.recordActivity(c, c.Params("shareId"), "REVOKE", "Mencabut link publik Dashboard KPI")
	return response.OK(c, "Link publik berhasil dicabut", nil)
}

func (h *Handler) Rotate(c *fiber.Ctx) error {
	created, err := h.service.Rotate(actorFromContext(c), c.Params("shareId"))
	if err != nil {
		return writeError(c, err)
	}
	h.recordActivity(c, created.ShareID, "ROTATE", "Merotasi token link publik Dashboard KPI")
	return response.OK(c, "Token link publik berhasil diperbarui. Link lama sudah tidak berlaku.", created)
}

func (h *Handler) recordActivity(c *fiber.Ctx, shareID, action, description string) {
	if h.activity == nil {
		return
	}
	_ = h.activity.Record(c.UserContext(), &logdomain.CreateActivityLogRequest{
		UserID: middleware.GetUserID(c), ActivityAction: action,
		TableAffected: "KPI_Public_Share", RecordID: shareID,
		ActivityDescription: description, IPAddress: c.IP(),
	})
}

func publicHeaders(c *fiber.Ctx) {
	c.Set("Cache-Control", "private, no-store")
	c.Set("Referrer-Policy", "no-referrer")
	c.Set("X-Robots-Tag", "noindex, nofollow")
}

func (h *Handler) Bootstrap(c *fiber.Ctx) error {
	publicHeaders(c)
	data, err := h.service.Bootstrap(c.Params("token"))
	if err != nil {
		return writeError(c, err)
	}
	return response.OK(c, "success", data)
}

type customQueryBody struct {
	WidgetID   string `json:"widget_id"`
	DrillLevel int    `json:"drill_level"`
}

func (h *Handler) CustomQuery(c *fiber.Ctx) error {
	publicHeaders(c)
	var body customQueryBody
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "Request body tidak valid", nil)
	}
	data, err := h.service.RunCustomQuery(c.Context(), c.Params("token"), strings.TrimSpace(body.WidgetID), body.DrillLevel)
	if err != nil {
		return writeError(c, err)
	}
	return response.OK(c, "success", data)
}

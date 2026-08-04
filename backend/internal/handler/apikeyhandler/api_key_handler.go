package apikeyhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/apikey"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/response"
)

type APIKeyHandler struct {
	uc  apikey.UseCase
	log *logger.Logger
}

func NewAPIKeyHandler(uc apikey.UseCase, log *logger.Logger) *APIKeyHandler {
	return &APIKeyHandler{uc: uc, log: log}
}

func getPlantID(c *fiber.Ctx) string {
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	userPlantID, _ := c.Locals("userPlantID").(string)

	if !isSuperAdmin {
		return userPlantID
	}

	return c.Query("plant_id")
}

func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	userPlantID, _ := c.Locals("userPlantID").(string)

	var req apikey.CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request payload", err.Error())
	}

	targetPlantID := ""
	if !isSuperAdmin {
		targetPlantID = userPlantID
	} else {
		if req.PlantID != nil && *req.PlantID != "" && *req.PlantID != "GLOBAL" && *req.PlantID != "ALL" {
			targetPlantID = *req.PlantID
		} else if qP := c.Query("plant_id"); qP != "" && qP != "GLOBAL" && qP != "ALL" {
			targetPlantID = qP
		}
	}

	res, err := h.uc.CreateKey(&req, userID, targetPlantID)
	if err != nil {
		h.log.Error("Failed to create API key", logger.Error(err))
		return response.InternalServerError(c, "Failed to create API key", err.Error())
	}

	return response.Created(c, "API Key created successfully. Copy the raw token now, it will not be displayed again.", res)
}

func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	plantID := getPlantID(c)
	keys, err := h.uc.ListKeys(userID, plantID)
	if err != nil {
		h.log.Error("Failed to list API keys", logger.Error(err))
		return response.InternalServerError(c, "Failed to list API keys", err.Error())
	}

	return response.OK(c, "success", keys)
}

func (h *APIKeyHandler) Revoke(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	keyID := c.Params("id")
	if keyID == "" {
		return response.BadRequest(c, "Key ID is required", nil)
	}

	if err := h.uc.RevokeKey(keyID, userID); err != nil {
		h.log.Error("Failed to revoke API key", logger.Error(err))
		return response.InternalServerError(c, "Failed to revoke API key", err.Error())
	}

	return response.OK(c, "API key revoked successfully", nil)
}

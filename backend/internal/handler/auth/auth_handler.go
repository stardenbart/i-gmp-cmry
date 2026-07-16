package auth

import (
	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

// AuthHandler handles authentication endpoints: login, logout, me.
type AuthHandler struct {
	authUC authdomain.AuthUseCase
}

func NewAuthHandler(authUC authdomain.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// Login godoc
// @Summary      User Login
// @Description  Authenticate user and return token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body interface{} true "Login credentials"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error  {
	var req authdomain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}

	ip := c.IP()
	device := c.Get("User-Agent")

	result, err := h.authUC.Login(&req, ip, device)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}
	return response.OK(c, "login successful", result)
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidate user session
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        login_log_id query string false "Login Log ID"
// @Success      200 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error  {
	userID := middleware.GetUserID(c)
	loginLogID := c.Query("login_log_id")

	if err := h.authUC.Logout(userID, loginLogID); err != nil {
		return response.InternalServerError(c, "logout failed", err.Error())
	}
	return response.OK(c, "logout successful", nil)
}

// Me godoc
// @Summary      Get current authenticated user
// @Description  Get current user profile
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *fiber.Ctx) error  {
	userID := middleware.GetUserID(c)
	info, err := h.authUC.Me(userID)
	if err != nil {
		return response.NotFound(c, "user not found")
	}
	return response.OK(c, "success", info)
}

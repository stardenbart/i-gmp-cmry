package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/config"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

// AuthHandler handles authentication endpoints: login, refresh, logout, me.
type AuthHandler struct {
	authUC authdomain.AuthUseCase
	cfg    *config.Config
}

func NewAuthHandler(authUC authdomain.AuthUseCase, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authUC: authUC, cfg: cfg}
}

// Login godoc
// @Summary      User Login
// @Description  Authenticate using username or email. On success the access
// @Description  and refresh tokens are set as httpOnly cookies — the
// @Description  response body only carries the user profile.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body interface{} true "Login credentials"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
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
	middleware.SetSessionCookies(c, h.cfg, result.AccessToken, result.AccessExpiresAt, result.RefreshToken, result.RefreshExpiresAt, result.CSRFToken)
	return response.OK(c, "login successful", fiber.Map{"user": result.User})
}

// Refresh godoc
// @Summary      Refresh session
// @Description  Rotates the refresh_token cookie into a fresh access + refresh
// @Description  token pair. Called silently by the frontend when a request
// @Description  gets a 401 for an expired access token.
// @Tags         Auth
// @Produce      json
// @Success      200 {object} response.APIResponse
// @Failure      401 {object} response.APIResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	raw := c.Cookies(middleware.CookieRefreshToken)

	result, err := h.authUC.Refresh(raw, c.IP(), c.Get("User-Agent"))
	if err != nil {
		// An invalid/expired/reused refresh token means the session is over
		// either way — clear whatever cookies remain so the browser doesn't
		// keep retrying a dead session.
		middleware.ClearSessionCookies(c, h.cfg)
		return response.Unauthorized(c, err.Error())
	}
	middleware.SetSessionCookies(c, h.cfg, result.AccessToken, result.AccessExpiresAt, result.RefreshToken, result.RefreshExpiresAt, result.CSRFToken)
	return response.OK(c, "token refreshed", fiber.Map{"user": result.User})
}

// Logout godoc
// @Summary      Logout
// @Description  Revokes the current refresh token and clears session cookies
// @Tags         Auth
// @Produce      json
// @Param        login_log_id query string false "Login Log ID"
// @Success      200 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	loginLogID := c.Query("login_log_id")
	raw := c.Cookies(middleware.CookieRefreshToken)

	if err := h.authUC.Logout(userID, loginLogID, raw); err != nil {
		return response.InternalServerError(c, "logout failed", err.Error())
	}
	middleware.ClearSessionCookies(c, h.cfg)
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
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	info, err := h.authUC.Me(userID)
	if err != nil {
		return response.NotFound(c, "user not found")
	}
	return response.OK(c, "success", info)
}

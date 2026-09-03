package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

// UserHandler handles user management endpoints.
type UserHandler struct {
	userUC authdomain.UserUseCase
}

func NewUserHandler(userUC authdomain.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// GetAll godoc
// @Summary      Get all users
// @Description  Get paginated list of users
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number"
// @Param        limit query int false "Limit per page"
// @Param        search query string false "Search term"
// @Param        role_id query string false "Role ID"
// @Param        department_id query string false "Department ID"
// @Success      200 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users [get]
func (h *UserHandler) GetAll(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	search := c.Query("search")
	roleID := c.Query("role_id")
	deptID := c.Query("department_id")
	plantID := c.Query("plant_id")

	userPlantID, _ := c.Locals("userPlantID").(string)
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	if !isSuperAdmin && userPlantID != "" {
		plantID = userPlantID
	}

	users, total, err := h.userUC.GetAll(p.Page, p.Limit, search, roleID, deptID, plantID)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch users", err.Error())
	}
	return response.Paginated(c, "success", users, total, p.Page, p.Limit)
}

// GetByID godoc
// @Summary      Get user by ID
// @Description  Get user details by ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.userUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, "user not found")
	}
	return response.OK(c, "success", user)
}

// Create godoc
// @Summary      Create user
// @Description  Create a new user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body interface{} true "User data"
// @Success      201 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users [post]
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req authdomain.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}

	userPlantID, _ := c.Locals("userPlantID").(string)
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	if !isSuperAdmin && userPlantID != "" {
		req.PlantID = &userPlantID
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	user, err := h.userUC.Create(&req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.Created(c, "user created successfully", user)
}

// Update godoc
// @Summary      Update user
// @Description  Update user details
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        body body interface{} true "User data"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req authdomain.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}

	// The route lets a user reach this handler for their OWN id without the
	// MOD-USR UPDATE permission (see router/v1/auth_routes.go), specifically
	// so anyone can edit their own name/email without needing user-management
	// rights. That bypass must never let the request also change fields that
	// grant privilege — role_id, user_status, department_id, or PIC mapping
	// scope — or any authenticated user could simply PUT their own id with
	// e.g. {"role_id":"ROLE-000"} and grant themselves Super Admin. Only an
	// admin going through the PermissionMiddleware-gated path (someone
	// else's id) may set those fields.
	if middleware.GetUserID(c) == id {
		req.RoleID = ""
		req.UserStatus = ""
		req.DepartmentID = ""
		req.PICKawasanIDs = nil
		req.PICKategori = nil
	}

	userPlantID, _ := c.Locals("userPlantID").(string)
	isSuperAdmin, _ := c.Locals("isSuperAdmin").(bool)
	if !isSuperAdmin && userPlantID != "" {
		req.PlantID = &userPlantID
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	user, err := h.userUC.Update(id, &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "user updated successfully", user)
}

// Delete godoc
// @Summary      Delete user
// @Description  Delete user by ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.userUC.Delete(id); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "user deleted successfully", nil)
}

// ChangePassword godoc
// @Summary      Change user password
// @Description  Change the password for a specific user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        body body interface{} true "Password data"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	id := c.Params("id")
	var req authdomain.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	if err := h.userUC.ChangePassword(id, &req); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "password changed successfully", nil)
}

// ResetPassword godoc
// @Summary      Admin reset password
// @Description  Admin resets user password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        body body interface{} true "Reset data"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /users/{id}/reset-password [post]
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	id := c.Params("id")
	var req authdomain.AdminResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	if err := h.userUC.AdminResetPassword(id, &req); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "password reset successfully", nil)
}

// ForgotPassword godoc
// @Summary      Forgot password
// @Description  Request a one-time password (OTP) for password reset
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body interface{} true "Email data"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Router       /auth/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var req authdomain.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	if err := h.userUC.ForgotPassword(&req); err != nil {
		if errors.Is(err, authdomain.ErrEmailNotRegistered) {
			return response.NotFound(c, "Email tidak terdaftar")
		}
		return response.InternalServerError(c, "Gagal mengirim kode OTP", err.Error())
	}
	return response.OK(c, "Kode OTP telah dikirim ke email Anda", nil)
}

// ResetPasswordWithOTP godoc
// @Summary      Reset password using OTP
// @Description  Verify the emailed OTP and set a new password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body authdomain.ResetPasswordWithOTPRequest true "OTP and new password"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Router       /auth/reset-password [post]
func (h *UserHandler) ResetPasswordWithOTP(c *fiber.Ctx) error {
	var req authdomain.ResetPasswordWithOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	if err := h.userUC.ResetPasswordWithOTP(&req); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "Password berhasil diubah. Silakan login dengan password baru.", nil)
}

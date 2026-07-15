package auth

import (
	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
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

func (h *UserHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	search := c.Query("search")
	roleID := c.Query("role_id")
	deptID := c.Query("department_id")

	users, total, err := h.userUC.GetAll(p.Page, p.Limit, search, roleID, deptID)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch users", err.Error())
	}
	return response.Paginated(c, "success", users, total, p.Page, p.Limit)
}

func (h *UserHandler) GetByID(c *fiber.Ctx) error  {
	id := c.Params("id")
	user, err := h.userUC.GetByID(id)
	if err != nil {
		return response.NotFound(c, "user not found")
	}
	return response.OK(c, "success", user)
}

func (h *UserHandler) Create(c *fiber.Ctx) error  {
	var req authdomain.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
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

func (h *UserHandler) Update(c *fiber.Ctx) error  {
	id := c.Params("id")
	var req authdomain.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
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

func (h *UserHandler) Delete(c *fiber.Ctx) error  {
	id := c.Params("id")
	if err := h.userUC.Delete(id); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "user deleted successfully", nil)
}

func (h *UserHandler) ChangePassword(c *fiber.Ctx) error  {
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

func (h *UserHandler) ResetPassword(c *fiber.Ctx) error  {
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

func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error  {
	var req authdomain.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	// Always respond OK regardless of email existence (anti-enumeration)
	_ = h.userUC.ForgotPassword(&req)
	return response.OK(c, "Jika email terdaftar, password sementara telah dikirim ke email Anda", nil)
}

package auth

import (
	"github.com/gin-gonic/gin"
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

func (h *UserHandler) GetAll(c *gin.Context) {
	p := pagination.FromQuery(c)
	search := c.Query("search")
	roleID := c.Query("role_id")
	deptID := c.Query("department_id")

	users, total, err := h.userUC.GetAll(p.Page, p.Limit, search, roleID, deptID)
	if err != nil {
		response.InternalServerError(c, "failed to fetch users", err.Error())
		return
	}
	response.Paginated(c, "success", users, total, p.Page, p.Limit)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userUC.GetByID(id)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.OK(c, "success", user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req authdomain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	user, err := h.userUC.Create(&req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Created(c, "user created successfully", user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req authdomain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	user, err := h.userUC.Update(id, &req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.OK(c, "user updated successfully", user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.userUC.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.OK(c, "user deleted successfully", nil)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	id := c.Param("id")
	var req authdomain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.userUC.ChangePassword(id, &req); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.OK(c, "password changed successfully", nil)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req authdomain.AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.userUC.AdminResetPassword(id, &req); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.OK(c, "password reset successfully", nil)
}

func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req authdomain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	// Always respond OK regardless of email existence (anti-enumeration)
	_ = h.userUC.ForgotPassword(&req)
	response.OK(c, "Jika email terdaftar, password sementara telah dikirim ke email Anda", nil)
}

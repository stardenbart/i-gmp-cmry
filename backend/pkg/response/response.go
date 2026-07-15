package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// APIResponse is the standard JSON response envelope for all API endpoints.
type APIResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Error      interface{} `json:"error,omitempty"`
}

// PaginatedResponse wraps paginated list results.
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int64       `json:"total_pages"`
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func OK(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(http.StatusOK).JSON(APIResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Message:    message,
		Data:       data,
	})
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(http.StatusCreated).JSON(APIResponse{
		Success:    true,
		StatusCode: http.StatusCreated,
		Message:    message,
		Data:       data,
	})
}

func BadRequest(c *fiber.Ctx, message string, err interface{}) error {
	return c.Status(http.StatusBadRequest).JSON(APIResponse{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Message:    message,
		Error:      err,
	})
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(http.StatusUnauthorized).JSON(APIResponse{
		Success:    false,
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	})
}

func Forbidden(c *fiber.Ctx, message string) error {
	return c.Status(http.StatusForbidden).JSON(APIResponse{
		Success:    false,
		StatusCode: http.StatusForbidden,
		Message:    message,
	})
}

func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(http.StatusNotFound).JSON(APIResponse{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Message:    message,
	})
}

func InternalServerError(c *fiber.Ctx, message string, err interface{}) error {
	return c.Status(http.StatusInternalServerError).JSON(APIResponse{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Message:    message,
		Error:      err,
	})
}

func Paginated(c *fiber.Ctx, message string, items interface{}, total int64, page, limit int) error {
	totalPages := total / int64(limit)
	if total%int64(limit) != 0 {
		totalPages++
	}
	return c.Status(http.StatusOK).JSON(APIResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Message:    message,
		Data: PaginatedResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

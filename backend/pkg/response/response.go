package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Message:    message,
		Data:       data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success:    true,
		StatusCode: http.StatusCreated,
		Message:    message,
		Data:       data,
	})
}

func BadRequest(c *gin.Context, message string, err interface{}) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Message:    message,
		Error:      err,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success:    false,
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success:    false,
		StatusCode: http.StatusForbidden,
		Message:    message,
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Message:    message,
	})
}

func InternalServerError(c *gin.Context, message string, err interface{}) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Message:    message,
		Error:      err,
	})
}

func Paginated(c *gin.Context, message string, items interface{}, total int64, page, limit int) {
	totalPages := total / int64(limit)
	if total%int64(limit) != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, APIResponse{
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

package pagination

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

// Params holds page/limit values parsed from query string.
type Params struct {
	Page   int
	Limit  int
	Offset int
}

// FromQuery extracts pagination params from a Gin context query string.
// Defaults: page=1, limit=10. Max limit is capped at 100.
func FromQuery(c *fiber.Ctx) Params {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = DefaultPage
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Params{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

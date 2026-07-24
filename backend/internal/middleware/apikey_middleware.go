package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/apikey"
	"github.com/monitoring-system/backend/pkg/response"
)

const ContextKeyAPIKey = "api_key_entity"

// APIKeyBearerMiddleware validates Bearer API token from Authorization header or query param.
func APIKeyBearerMiddleware(uc apikey.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			return c.Next()
		}

		tokenStr := ""

		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr = parts[1]
			}
		}

		if tokenStr == "" {
			tokenStr = c.Query("api_key")
		}
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}
		if tokenStr == "" {
			tokenStr = c.Query("key")
		}

		if tokenStr == "" {
			return response.Unauthorized(c, "API Key is missing. Please provide header 'Authorization: Bearer MAKEY_...' or query parameter '?api_key=MAKEY_...'")
		}

		keyEntity, err := uc.ValidateAndConsumeKey(tokenStr)
		if err != nil {
			return response.Unauthorized(c, err.Error())
		}

		c.Locals(ContextKeyAPIKey, keyEntity)
		return c.Next()
	}
}

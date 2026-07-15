package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CORSMiddleware sets CORS headers based on allowed origins from config.
func CORSMiddleware(allowedOrigins string) fiber.Handler {
	origins := strings.Split(allowedOrigins, ",")
	allowedMap := make(map[string]bool)
	for _, o := range origins {
		allowedMap[strings.TrimSpace(o)] = true
	}

	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		if allowedMap[origin] || allowedOrigins == "*" {
			c.Set("Access-Control-Allow-Origin", origin)
		}

		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Max-Age", "86400")
		c.Set("Vary", "Origin")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		// Set response time header for debugging
		start := time.Now()
		err := c.Next()
		c.Set("X-Response-Time", time.Since(start).String())
		return err
	}
}

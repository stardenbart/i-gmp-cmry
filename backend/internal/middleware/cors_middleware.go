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

		// Auth now rides on cookies (Access-Control-Allow-Credentials below),
		// so this check is a real access-control boundary, not just a CORS
		// header for convenience: reflecting an origin we didn't actually
		// allow would let any website read authenticated API responses via a
		// credentialed fetch(), because the browser attaches session cookies
		// automatically regardless of which site's JS made the request.
		// Only ever echo back an origin that matched the allowlist (or the
		// ngrok dev-tunnel suffixes) — never fall back to reflecting
		// whatever the caller sent.
		if allowedOrigins == "*" || allowedMap[origin] || strings.HasSuffix(origin, ".ngrok-free.dev") || strings.HasSuffix(origin, ".ngrok.io") {
			c.Set("Access-Control-Allow-Origin", origin)
		}

		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, Access-Control-Request-Private-Network")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Allow-Private-Network", "true")
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

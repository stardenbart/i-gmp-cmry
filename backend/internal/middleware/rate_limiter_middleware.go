package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// AuthRateLimiter restricts sensitive authentication endpoints (Login/Reset Pass)
// to maximum 5 requests per minute per client IP address (Brute Force Protection).
func AuthRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			ip := c.IP()
			if ip == "" {
				ip = c.Get("X-Forwarded-For")
			}
			return ip
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak percobaan login. Harap tunggu 1 menit sebelum mencoba kembali.",
			})
		},
	})
}

// APIRateLimiter restricts general API endpoints to maximum 100 requests per minute per IP.
func APIRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			ip := c.IP()
			if ip == "" {
				ip = c.Get("X-Forwarded-For")
			}
			return ip
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Batas request API terlampaui. Harap perlambat request Anda.",
			})
		},
	})
}

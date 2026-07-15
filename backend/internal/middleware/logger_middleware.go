package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/logger"
	"go.uber.org/zap"
)

// LoggerMiddleware logs every incoming HTTP request using Zap structured logger.
func LoggerMiddleware(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		query := string(c.Request().URI().QueryString())

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		fields := []zap.Field{
			logger.String("method", c.Method()),
			logger.String("path", path),
			logger.String("query", query),
			logger.Int("status", status),
			logger.String("ip", c.IP()),
			logger.String("latency", latency.String()),
			logger.String("user_agent", c.Get("User-Agent")),
		}

		if userID := GetUserID(c); userID != "" {
			fields = append(fields, logger.String("user_id", userID))
		}

		if status >= 500 {
			log.Error("server error", fields...)
		} else if status >= 400 {
			log.Warn("client error", fields...)
		} else {
			log.Info("request completed", fields...)
		}
		return err
	}
}

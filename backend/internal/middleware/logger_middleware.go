package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/pkg/logger"
	"go.uber.org/zap"
)

// LoggerMiddleware logs every incoming HTTP request using Zap structured logger.
func LoggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.String("query", query),
			logger.Int("status", status),
			logger.String("ip", c.ClientIP()),
			logger.String("latency", latency.String()),
			logger.String("user_agent", c.Request.UserAgent()),
		}

		if userID := GetUserID(c); userID != "" {
			fields = append(fields, logger.String("user_id", userID))
		}

		if len(c.Errors) > 0 {
			log.Error("request error", logger.Any("errors", c.Errors.Errors()))
		} else if status >= 500 {
			log.Error("server error", fields...)
		} else if status >= 400 {
			log.Warn("client error", fields...)
		} else {
			log.Info("request completed", fields...)
		}
	}
}

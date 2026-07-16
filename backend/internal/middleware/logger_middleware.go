package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
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

// httpMethodToAction converts HTTP methods to CRUD action names.
func httpMethodToAction(method string) string {
	switch strings.ToUpper(method) {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "READ"
	}
}

// pathToTable extracts a simple table/module name from the URL path.
// e.g. "/api/v1/issues/ISU-001" → "Issue"
func pathToTable(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for i, seg := range segments {
		if seg == "v1" && i+1 < len(segments) {
			raw := segments[i+1]
			// Normalize known route prefixes to table names
			tableMap := map[string]string{
				"issues":           "Issue",
				"issues/photos":    "IssuePhoto",
				"inspections":      "Inspection",
				"master":           "Master",
				"users":            "User",
				"roles":            "Role",
				"logs":             "Log",
				"upload":           "Upload",
				"auth":             "Auth",
			}
			if name, ok := tableMap[raw]; ok {
				return name
			}
			// Capitalize the first letter as fallback
			if len(raw) > 0 {
				return strings.ToUpper(raw[:1]) + raw[1:]
			}
		}
	}
	return "Unknown"
}

// ActivityLogMiddleware intercepts every authenticated request and records a log entry
// asynchronously to the Activity_Log table (via Kafka → DB and direct DB write).
// Only successful responses (2xx) are logged to avoid noise from client errors.
func ActivityLogMiddleware(actLogUC logdomain.ActivityLogUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		status := c.Response().StatusCode()
		// Only log successful mutating operations (2xx on non-GET requests) and all GETs for audit trail
		if status < 200 || status >= 300 {
			return err
		}

		userID := GetUserID(c)
		if userID == "" {
			return err // skip unauthenticated requests
		}

		method := c.Method()
		action := httpMethodToAction(method)
		table := pathToTable(c.Path())
		recordID := c.Params("id")
		desc := fmt.Sprintf("[%s] %s", method, c.Path())

		// Fire-and-forget async logging — must not block the HTTP response
		go func() {
			_ = actLogUC.Record(context.Background(), &logdomain.CreateActivityLogRequest{
				UserID:              userID,
				ActivityAction:      action,
				TableAffected:       table,
				RecordID:            recordID,
				ActivityDescription: desc,
				IPAddress:           c.IP(),
			})
		}()

		return err
	}
}


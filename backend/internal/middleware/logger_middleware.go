package middleware

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/pkg/logger"
	"go.uber.org/zap"
)

// base64ImageRegex mencocokkan data URL base64 gambar di dalam JSON string
// Contoh: "data:image/webp;base64,UklGR..." (bisa ratusan KB)
var base64ImageRegex = regexp.MustCompile(`data:image/[a-zA-Z+]+;base64,[A-Za-z0-9+/=]{100,}`)

// stripBase64FromJSON mengganti semua base64 data URL dengan placeholder ringkas
// sehingga Activity_Log tidak menyimpan data biner besar
func stripBase64FromJSON(s string) string {
	return base64ImageRegex.ReplaceAllStringFunc(s, func(match string) string {
		return fmt.Sprintf("[base64_image:%d_bytes]", len(match))
	})
}


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
				"issues":        "Issue",
				"issues/photos": "IssuePhoto",
				"inspections":   "Inspection",
				"master":        "Master",
				"users":         "User",
				"roles":         "Role",
				"logs":          "Log",
				"upload":        "Upload",
				"auth":          "Auth",
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
		// Only log successful operations
		if status < 200 || status >= 300 {
			return err
		}

		method := c.Method()
		// Only log mutating operations (POST, PUT, PATCH, DELETE) for audit trail.
		// Skip read-only GET/HEAD/OPTIONS requests to eliminate useless DB I/O thrashing.
		if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
			return err
		}

		path := c.Path()
		// Skip audit logging for high-frequency background polling/sync/heartbeat/lock/permission paths to prevent DB I/O thrashing
		if strings.Contains(path, "/sync") ||
			strings.Contains(path, "/heartbeat") ||
			strings.Contains(path, "/notifications") ||
			strings.Contains(path, "/health") ||
			strings.Contains(path, "/metrics") ||
			strings.Contains(path, "/lock") ||
			strings.Contains(path, "/pic-mappings") ||
			strings.Contains(path, "/permissions") {
			return err
		}

		userID := GetUserID(c)
		if userID == "" {
			return err // skip unauthenticated requests
		}

		// Capture all values from Fiber ctx BEFORE the goroutine.
		// Fiber recycles ctx after the handler returns, so accessing c inside
		// a goroutine causes a nil pointer dereference (panic: SIGSEGV).
		action := httpMethodToAction(method)
		table := pathToTable(c.Path())
		recordID := c.Params("id")
		if recordID == "" {
			recordID = c.Query("id")
		}

		query := string(c.Request().URI().QueryString())
		desc := fmt.Sprintf("[%s] %s", method, path)
		if query != "" {
			desc += "?" + query
		}

		ipAddress := c.IP() // must be read here, not inside the goroutine

		// Capture request payload body
		// For POST/PUT/PATCH/DELETE: use the raw JSON body
		// For GET: build a descriptive JSON payload from path, query params, and headers
		var reqBody string
		contentType := c.Get("Content-Type")
		bodyBytes := c.Body()

		if strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "image/") || strings.Contains(contentType, "octet-stream") {
			reqBody = fmt.Sprintf("[file upload payload: %s, size: %d bytes]", contentType, len(bodyBytes))
		} else if len(bodyBytes) > 0 {
			rawStr := string(bodyBytes)
			// Strip base64 data URLs sebelum logging — bisa 500KB+ per foto WebP
			// Ganti dengan placeholder ringkas agar INSERT Activity_Log tidak lambat
			rawStr = stripBase64FromJSON(rawStr)
			if len(rawStr) > 2000 {
				rawStr = rawStr[:2000] + "... (truncated)"
			}
			if !utf8.ValidString(rawStr) {
				rawStr = strings.ToValidUTF8(rawStr, "?")
			}
			reqBody = rawStr
		} else {
			// For GET (READ) — build a payload from query information
			reqPayload := fmt.Sprintf(`{
  "method": "%s",
  "path": "%s",
  "query_params": "%s",
  "user_agent": "%s",
  "ip": "%s"
}`, method, path, query, c.Get("User-Agent"), ipAddress)
			reqBody = reqPayload
		}

		// Fire-and-forget async logging — must not block the HTTP response
		go func() {
			_ = actLogUC.Record(context.Background(), &logdomain.CreateActivityLogRequest{
				UserID:              userID,
				ActivityAction:      action,
				TableAffected:       table,
				RecordID:            recordID,
				NewValue:            reqBody,
				ActivityDescription: desc,
				IPAddress:           ipAddress,
			})
		}()

		return err
	}
}

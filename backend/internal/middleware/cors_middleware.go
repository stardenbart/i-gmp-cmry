package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware sets CORS headers based on allowed origins from config.
func CORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	allowedMap := make(map[string]bool)
	for _, o := range origins {
		allowedMap[strings.TrimSpace(o)] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowedMap[origin] || allowedOrigins == "*" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Set response time header for debugging
		start := time.Now()
		c.Next()
		c.Header("X-Response-Time", time.Since(start).String())
	}
}

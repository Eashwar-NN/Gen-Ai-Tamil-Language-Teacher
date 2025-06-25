package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a middleware that logs request details
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		logEntry := fmt.Sprintf("[%s] %s | %d | %v | %s | %s",
			method, path, statusCode, latency, clientIP, query)

		switch {
		case statusCode >= 500:
			fmt.Printf("\x1b[31m%s\x1b[0m\n", logEntry) // Red for server errors
		case statusCode >= 400:
			fmt.Printf("\x1b[33m%s\x1b[0m\n", logEntry) // Yellow for client errors
		case statusCode >= 300:
			fmt.Printf("\x1b[36m%s\x1b[0m\n", logEntry) // Cyan for redirects
		default:
			fmt.Printf("\x1b[32m%s\x1b[0m\n", logEntry) // Green for success
		}

		// Log errors if any
		if len(c.Errors) > 0 {
			fmt.Printf("\x1b[31mErrors: %v\x1b[0m\n", c.Errors.JSON())
		}
	}
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
} 
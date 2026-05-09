// Package middlewares provides Gin middleware for cross-cutting concerns.
package middlewares

import (
	"log/slog"
	"time"

	"github.com/code-execution-engine/pkg/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger returns a Gin middleware that logs every incoming HTTP request
// with structured fields: request_id, method, path, status, latency, client_ip,
// and user_agent. It also injects a unique request ID into the context and the
// response header (X-Request-ID) for distributed tracing.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- Before request ---
		start := time.Now()
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store request ID in context for downstream handlers/services.
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Process the request
		c.Next()

		// --- After request ---
		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Int("body_bytes", c.Writer.Size()),
		}

		switch {
		case status >= 500:
			telemetry.Error("Server error", attrs...)
		case status >= 400:
			telemetry.Warn("Client error", attrs...)
		default:
			telemetry.Info("Request completed", attrs...)
		}
	}
}

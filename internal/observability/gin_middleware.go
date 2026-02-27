package observability

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDKey = "request_id"

func RequestMetricsAndLoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(requestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := c.Writer.Status()
		latency := time.Since(start)

		RecordHTTPRequest(c.Request.Method, path, status, latency)
		logger.Info(
			"http_request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func RequestIDFromGin(c *gin.Context) string {
	v, ok := c.Get(requestIDKey)
	if !ok {
		return ""
	}
	id, ok := v.(string)
	if !ok {
		return ""
	}
	return id
}

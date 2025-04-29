package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"log/slog"
	"project_reference/infrastructure/logger"
	"time"
)

func LoggerMiddleware() gin.HandlerFunc {
	const (
		requestIdHeader     = "X-Request-Id"
		requestRealIpHeader = "X-Real-Ip"
	)

	return func(c *gin.Context) {
		start := time.Now()
		query := c.Request.URL.RawQuery
		path := c.Request.URL.Path
		if query != "" {
			path += "?" + query
		}

		requestId := c.GetHeader(requestIdHeader)
		if requestId == "" {
			requestId = xid.New().String()
		}

		requestIp := c.GetHeader(requestRealIpHeader)
		if requestIp == "" {
			requestIp = c.ClientIP()
		}

		ctx := logger.WithTraceID(c.Request.Context(), requestId)

		log := logger.Logger(ctx).With(
			"server_request", slog.GroupValue(
				slog.String("request_id", requestId),
				slog.String("method", c.Request.Method),
				slog.String("path", path),
				slog.String("ip", requestIp),
				slog.String("user_agent", c.Request.UserAgent()),
			),
		)

		c.Set("requestId", requestId)
		c.Set("requestIp", requestIp)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start).Milliseconds()

		log.With(
			"server_response", slog.GroupValue(
				slog.Int("status", status),
				slog.Float64("latency", float64(latency)*0.001),
			),
		).Info("request")
	}
}

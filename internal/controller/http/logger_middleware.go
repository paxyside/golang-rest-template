package http

import (
	"golang-template/internal/domain/logger"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func LoggerMiddleware(log logger.Loggerer) gin.HandlerFunc {
	const (
		requestIDHeader     = "X-Request-Id"
		requestRealIPHeader = "X-Real-Ip"
	)

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if query := c.Request.URL.RawQuery; query != "" {
			path += "?" + query
		}

		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = xid.New().String()
		}

		requestIP := c.GetHeader(requestRealIPHeader)
		if requestIP == "" {
			requestIP = c.ClientIP()
		}

		logWithCtx := log.With(slog.Group("request_info",
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("ip", requestIP),
			slog.String("user_agent", c.Request.UserAgent()),
		))

		c.Writer.Header().Set(requestIDHeader, requestID)

		c.Next()

		statusCode := c.Writer.Status()
		latency := time.Since(start).Milliseconds()

		logWithCtx = logWithCtx.With(slog.Group("response",
			slog.Int("status", statusCode),
			slog.Int64("latency_ms", latency),
		))

		switch {
		case statusCode >= http.StatusInternalServerError:
			logWithCtx.Error("request failed")
		case statusCode >= http.StatusBadRequest:
			logWithCtx.Warn("request failed")
		default:
			logWithCtx.Info("request")
		}
	}
}

package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"log/slog"
	"time"
)

func LoggerMiddleware() gin.HandlerFunc {
	const (
		requestIdHeader     = "X-Request-Id"
		requestRealIpHeader = "X-Real-Ip"
	)

	return func(c *gin.Context) {
		var (
			start     = time.Now()
			query     = c.Request.URL.RawQuery
			path      = c.Request.URL.Path
			requestId string
			requestIp string
		)

		if query != "" {
			path = c.Request.URL.Path + "?" + query
		}

		if requestId = c.GetHeader(requestIdHeader); requestId == "" {
			requestId = xid.New().String()
		}

		if requestIp = c.GetHeader(requestRealIpHeader); requestIp == "" {
			requestIp = c.ClientIP()
		}

		l := slog.Default()
		l = l.With("server_request", slog.GroupValue(
			slog.String("request_id", requestId),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("ip", requestIp),
			slog.String("user_agent", c.Request.UserAgent()),
		))

		c.Writer.Header().Set(requestIdHeader, requestId)
		c.Set("requestId", requestId)
		c.Set("requestIp", requestIp)

		c.Next()

		responseStatusCode := c.Writer.Status()
		latency := time.Since(start).Milliseconds()

		l = l.With("server_response", slog.GroupValue(
			slog.Int("status", responseStatusCode),
			slog.Float64("latency", float64(latency)*0.001),
		))

		l.Info("request")
	}
}

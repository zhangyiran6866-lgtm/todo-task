package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	applog "todotask/backend/pkg/logger"
)

const CtxRequestIDKey = "request_id"

// RequestLog injects request_id and emits a structured access log for each HTTP request.
func RequestLog(base *zap.Logger) gin.HandlerFunc {
	if base == nil {
		base = zap.NewNop()
	}

	return func(c *gin.Context) {
		start := time.Now()

		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set(CtxRequestIDKey, requestID)
		c.Request = c.Request.WithContext(applog.WithRequestID(c.Request.Context(), requestID))

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		fields := []zap.Field{
			zap.String(applog.FieldModule, "http"),
			zap.String(applog.FieldAction, "request"),
			zap.String(applog.FieldRequestID, requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("route", c.FullPath()),
			zap.Int("status_code", statusCode),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if userID, ok := c.Get(CtxUserIDKey); ok {
			if userIDStr, ok := userID.(string); ok && userIDStr != "" {
				fields = append(fields, zap.String(applog.FieldUserID, userIDStr))
				c.Request = c.Request.WithContext(applog.WithUserID(c.Request.Context(), userIDStr))
			}
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		reqLogger := base.With(fields...)
		switch {
		case statusCode >= 500:
			reqLogger.Error("http request completed")
		case statusCode >= 400:
			reqLogger.Warn("http request completed")
		default:
			reqLogger.Info("http request completed")
		}
	}
}

func newRequestID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

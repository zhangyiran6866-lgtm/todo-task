package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"todotask/backend/pkg/config"
	"todotask/backend/pkg/jwt"
	applog "todotask/backend/pkg/logger"
	"todotask/backend/pkg/response"
)

const CtxUserIDKey = "user_id"

// JWTAuth returns a gin middleware for validating the access token
func JWTAuth(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少 Authorization 请求头")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Authorization 格式不正确")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := jwt.ParseToken(tokenStr, cfg.AccessSecret)
		if err != nil {
			response.Unauthorized(c, "访问令牌无效或已过期")
			c.Abort()
			return
		}

		// Set the UserID in the context for downstream handlers
		userID := claims.UserID.Hex()
		c.Set(CtxUserIDKey, userID)
		c.Request = c.Request.WithContext(applog.WithUserID(c.Request.Context(), userID))
		c.Next()
	}
}

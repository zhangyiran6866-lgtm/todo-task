package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"todotask/backend/pkg/config"
	"todotask/backend/pkg/jwt"
	"todotask/backend/pkg/response"
)

const CtxUserIDKey = "user_id"

// JWTAuth returns a gin middleware for validating the access token
func JWTAuth(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := jwt.ParseToken(tokenStr, cfg.AccessSecret)
		if err != nil {
			response.Unauthorized(c, "invalid or expired access token")
			c.Abort()
			return
		}

		// Set the UserID in the context for downstream handlers
		c.Set(CtxUserIDKey, claims.UserID.Hex())
		c.Next()
	}
}

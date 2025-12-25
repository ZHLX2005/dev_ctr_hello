package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// AuthHeader 认证 HTTP 头名称
	AuthHeader = "Authorization"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	token string
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(token string) *AuthMiddleware {
	return &AuthMiddleware{
		token: token,
	}
}

// NewAuthMiddlewareFromEnv 从环境变量创建认证中间件
func NewAuthMiddlewareFromEnv() *AuthMiddleware {
	token := getEnv("AUTH_TOKEN", "your-secret-token-change-me")
	return NewAuthMiddleware(token)
}

// RequireAuth 需要认证的中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization 头
		authHeader := c.GetHeader(AuthHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查 Bearer token 格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format, use: Bearer <token>",
			})
			c.Abort()
			return
		}

		// 提取 token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != m.token {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

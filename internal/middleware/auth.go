package middleware

import (
	"context"
	"net/http"
	"strings"

	"DistanceBack_v1/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authorization header is required",
			})
			return
		}

		// 解析 token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid authorization header format",
			})
			return
		}

		token := parts[1]

		// 验证 Firebase token
		firebaseToken, err := auth.VerifyIDToken(context.Background(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid token",
			})
			return
		}

		// 将用户信息存储在上下文中
		c.Set("firebase_user", firebaseToken)
		c.Set("user_id", firebaseToken.UID)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}

		token := parts[1]

		firebaseToken, err := auth.VerifyIDToken(context.Background(), token)
		if err != nil {
			c.Next()
			return
		}

		c.Set("firebase_user", firebaseToken)
		c.Set("user_id", firebaseToken.UID)

		c.Next()
	}
}

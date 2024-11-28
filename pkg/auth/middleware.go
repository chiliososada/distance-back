package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AuthUserKey 是上下文中存储认证用户的键
const AuthUserKey = "auth_user"

// TokenFromHeader 从请求头中提取token
func TokenFromHeader(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

// AuthMiddleware Firebase认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := TokenFromHeader(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrAuthentication)
			return
		}

		// 验证token
		decodedToken, err := VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			logger.Error("Failed to verify Firebase ID token",
				logger.Any("error", err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrAuthentication)
			return
		}

		// 将用户信息存储在上下文中
		authUser := NewAuthUserFromToken(decodedToken) // 使用 NewAuthUserFromToken 创建认证用户
		c.Set(AuthUserKey, authUser)                   // 存储 AuthUser 实例
		c.Next()
	}
}

// OptionalAuthMiddleware 可选的认证中间件（不强制要求认证）
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := TokenFromHeader(c)
		if err != nil {
			c.Next()
			return
		}

		decodedToken, err := VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			logger.Warn("Failed to verify optional Firebase ID token",
				logger.Any("error", err))
			c.Next()
			return
		}

		c.Set(AuthUserKey, decodedToken)
		c.Next()
	}
}

// GetAuthUser 从上下文中获取认证用户
func GetAuthUser(ctx context.Context) (*AuthUser, error) {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if user, exists := ginCtx.Get(AuthUserKey); exists {
			if authUser, ok := user.(*AuthUser); ok {
				return authUser, nil
			}
		}
	}
	return nil, fmt.Errorf("unauthorized")
}

// RequireRoles 角色要求中间件
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetAuthUser(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrAuthentication)
			return
		}

		// 检查用户是否具有所需角色
		hasRole := false
		for _, role := range roles {
			if user.HasRole(role) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, errors.ErrAuthorization)
			return
		}

		c.Next()
	}
}

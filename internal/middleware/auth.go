package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/chiliososada/distance-back/pkg/auth"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
)

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie("Authorization")
		if err != nil {
			logger.Error("session err: ", logger.Err(err))
			c.AbortWithError(http.StatusUnauthorized, errors.ErrUnauthorized)
		}

		logger.Info("session ", logger.String("session", session))

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

package middleware

import (
	"net/http"

	"github.com/chiliososada/distance-back/pkg/auth"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
)

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie("Authorization")
		if err != nil {
			if err == http.ErrNoCookie {
				c.AbortWithStatus(http.StatusUnauthorized)
			} else {
				logger.Error("Find Authorization Cookie Failed", logger.Err(err))
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}

		token, err := auth.VeirfySessionCookie(c.Request.Context(), session)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = auth.SetSessionDataInContext(c, token.UID, session)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
		} else {
			c.Next()
		}
		return
	}
}

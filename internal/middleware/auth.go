package middleware

import (
	"fmt"
	"net/http"

	"github.com/chiliososada/distance-back/pkg/auth"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
)

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie("Authorization")
		fmt.Printf("session: %v\n", session)
		if err != nil {
			fmt.Printf("err: %v\n", err.Error())
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

		sessionData, err := auth.GetSessionData(token.UID, session)
		fmt.Printf("session data %v\n", sessionData)
		if err != nil {
			fmt.Printf("err: %v\n", err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
		} else if sessionData != nil {
			auth.SetSessionInContext(c, sessionData)
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		return
	}
}

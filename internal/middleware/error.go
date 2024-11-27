package middleware

import (
	"net/http"
	"runtime/debug"

	"DistanceBack_v1/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误堆栈
				logger.Error("Panic recovered",
					logger.Any("error", err),
					logger.String("stack", string(debug.Stack())),
				)

				// 返回 500 错误
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal Server Error",
				})
			}
		}()

		c.Next()
	}
}

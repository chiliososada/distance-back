package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 将新的上下文替换到请求中
		c.Request = c.Request.WithContext(ctx)

		// 创建完成通道
		done := make(chan bool, 1)

		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-ctx.Done():
			// 超时处理
			c.AbortWithStatusJSON(504, gin.H{
				"code":    504,
				"message": "request timeout",
			})
		case <-done:
			// 请求正常完成
		}
	}
}

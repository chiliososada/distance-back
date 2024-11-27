package middleware

import (
	"bytes"
	"io/ioutil"
	"time"

	"DistanceBack_v1/pkg/logger"

	"github.com/gin-gonic/gin"
)

// RequestLoggerConfig 请求日志配置
type RequestLoggerConfig struct {
	SkipPaths []string // 不记录日志的路径
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// RequestLogger 请求日志中间件
func RequestLogger(config RequestLoggerConfig) gin.HandlerFunc {
	skip := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skip[path] = true
	}

	return func(c *gin.Context) {
		// 检查是否跳过日志记录
		if skip[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器以捕获响应体
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 记录日志
		if raw != "" {
			path = path + "?" + raw
		}

		// 获取用户ID
		userID, _ := c.Get("user_id")

		logger.Info("Request completed",
			logger.String("path", path),
			logger.String("method", c.Request.Method),
			logger.Int("status", c.Writer.Status()),
			logger.Duration("latency", latency),
			logger.String("ip", c.ClientIP()),
			logger.Any("user_id", userID),
			logger.String("request", string(requestBody)),
			logger.String("response", w.body.String()),
		)
	}
}

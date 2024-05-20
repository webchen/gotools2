package middleware

import (
	"io"
	"net/http"
	"time"

	"github.com/webchen/gotools2/src/util/logs"

	"github.com/gin-gonic/gin"
)

func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.Request.Host
		// post 数据
		b := make([]byte, 0)

		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			b, _ = io.ReadAll(c.Request.Body)
		}

		logs.WebAccess("| %3d | %13v | %15s | %s | %s | %s",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
			string(b),
		)
	}
}

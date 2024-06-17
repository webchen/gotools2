package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/webchen/gotools2/src/util/logs"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 读取响应数据
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// post 数据
		b := make([]byte, 0)

		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			b, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
		}

		// 初始化bodyLogWriter
		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

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
		clientIP := c.ClientIP()

		responseBody := blw.body.String()

		logs.WebAccess("| %3d | %13v | %15s | %s | %s | %s | %s ",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
			string(b),
			responseBody,
		)
	}
}

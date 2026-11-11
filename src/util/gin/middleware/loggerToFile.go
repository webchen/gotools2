package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/webchen/gotools2/src/fun"
	"github.com/webchen/gotools2/src/util/logs"

	"github.com/gin-gonic/gin"
)

// 响应缓冲预留上限，防止异常 Content-Length 触发巨额分配
const maxCaptureGrow = 4 << 20

type bodyLogWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	tryGrow bool
}

// Write 读取响应数据；首次写入时按 Content-Length 预留容量，避免大响应反复扩容拷贝
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if w.tryGrow {
		w.tryGrow = false
		if n, err := strconv.Atoi(w.ResponseWriter.Header().Get("Content-Length")); err == nil && n > 0 && n <= maxCaptureGrow {
			w.body.Grow(n)
		}
	}
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// getTraceId 取 traceid：优先 query 的 traceid，其次请求头 x-b3-traceid（header 取值大小写不敏感）
func getTraceId(c *gin.Context) string {
	if traceId := c.Query("traceid"); traceId != "" {
		return traceId
	}
	return c.GetHeader("x-b3-traceid")
}

func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 访问日志关闭时，不做请求体读取、响应体缓冲等采集动作，直接放行
		if !logs.WebAccessEnabled() {
			c.Next()
			return
		}

		// 探针请求不写访问日志，也不值得为它读 body、缓冲响应
		ua := c.Request.UserAgent()
		if strings.Contains(ua, "kube-probe/") || strings.Contains(ua, "SLBHealthCheck") {
			c.Next()
			return
		}

		// 开始时间
		startTime := time.Now()

		// post 数据
		var b []byte
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			b, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
		}

		// 初始化bodyLogWriter
		blw := &bodyLogWriter{
			body:           new(bytes.Buffer),
			tryGrow:        true,
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
		// trace_id：优先 query 的 traceid，其次请求头 x-b3-traceid
		traceId := getTraceId(c)

		p, _ := fun.ZhToUnicode(b)

		// 整行一次组装：请求体与响应体直接写字节，避免 %s 造成的整段字符串拷贝；
		// 分隔与列顺序与原来的 "| %3d | %13v | %15s | %s | %s | %s | %s | %s " 逐字节等价
		var sb strings.Builder
		sb.Grow(len(p) + blw.body.Len() + 128)
		fmt.Fprintf(&sb, "| %3d | %13v | %15s | %s | %s | %s | ",
			statusCode,
			latencyTime,
			clientIP,
			traceId,
			reqMethod,
			reqUri,
		)
		sb.Write(p)
		sb.WriteString(" | ")
		sb.Write(blw.body.Bytes())
		sb.WriteString(" ")

		logs.WebAccessLine(sb.String())
	}
}

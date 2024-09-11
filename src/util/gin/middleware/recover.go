package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/webchen/gotools2/src/fun"
	"github.com/webchen/gotools2/src/util/logs"
	"github.com/webchen/gotools2/src/util/model2"
)

func HttpRecover(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			msg := "系统内部错误，请联系管理员"
			var data interface{}
			if e, ok := err.(*model2.PanicMessage); ok {
				msg = e.Message
				data = e.Data
			}
			if e, ok := err.(string); ok {
				msg = e
			}
			logs.Error(msg, data)
			fun.SendErrorResponse(c, msg)
			c.Abort()
		}
	}()
	c.Next()
}

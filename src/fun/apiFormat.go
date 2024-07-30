package fun

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ApiFormat 对外API格式
func ApiFormat(code int, data interface{}, message string) (m map[string]interface{}) {
	m = make(map[string]interface{})
	m["code"] = code
	if strings.TrimSpace(message) == "" {
		message = "success"
	}
	m["message"] = message
	if data == nil {
		data = make(map[string]interface{})
	}
	m["data"] = data
	m["timestamp"] = time.Now().Unix()

	return m
}

// ApiFormatSuccess  返回成功的JSON字符串
func ApiFormatSuccess(data interface{}, message string) (m map[string]interface{}) {
	return ApiFormat(1, data, message)
}

// ApiFormatFail  返回失败的JSON字符串
func ApiFormatFail(message string) (m map[string]interface{}) {
	if strings.TrimSpace(message) == "" {
		message = "fail"
	}
	return ApiFormat(0, nil, message)
}

func SendReponse(c *gin.Context, err error, data interface{}, msg string) {
	code := 1
	if err != nil {
		code = 0
		if msg == "" {
			msg = err.Error()
		}
	}
	if data == nil {
		data = make(map[string]interface{}, 0)
	}
	m := ApiFormat(code, data, msg)
	c.JSON(200, m)
	c.Abort()
}

func SendErrorResponse(c *gin.Context, errMsg string) {
	SendReponse(c, fmt.Errorf(errMsg), nil, "")
}

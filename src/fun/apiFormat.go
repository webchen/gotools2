package fun

import (
	"strings"
	"time"
)

// ApiFormat 对外API格式
func ApiFormat(code uint8, data interface{}, message string) (m map[string]interface{}) {
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

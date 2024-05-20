package fun

import (
	"gotools2/src/base/conf"
	"time"

	"gotools2/src/base/jsontool"

	"github.com/valyala/fasthttp"
)

var fasthttpClient = &fasthttp.Client{ReadTimeout: 5 * time.Second}

func doHTTP2(method string, url string, jsonMap map[string]interface{}) string {

	// 初始化请求与响应
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		// 用完需要释放资源
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
		//fasthttpClient.CloseIdleConnections()
	}()
	requestBody := []byte("")
	if method == "POST" {
		// 默认是application/x-www-form-urlencoded
		req.Header.SetContentType("application/json")
		req.Header.SetMethod(method)
		strs := jsontool.MarshalToString(jsonMap)
		requestBody = []byte(strs)
		req.SetBody(requestBody)
	}
	req.SetRequestURI(url)
	t, _ := time.ParseDuration(conf.GetConfig("system.http.queryTimeOut", "3s").(string))
	err := fasthttpClient.DoTimeout(req, resp, time.Second*t)
	if err != nil {
		logData := make(map[string]string)
		logData["url"] = url
		logData["error"] = err.Error()
		logData["method"] = method
		logData["postContent"] = string(requestBody)
	}

	b := resp.Body()
	return string(b)
}

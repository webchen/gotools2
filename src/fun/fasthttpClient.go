package fun

import (
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/webchen/gotools2/src/base/conf"

	"github.com/webchen/gotools2/src/base/jsontool"

	"github.com/valyala/fasthttp"
)

var fasthttpClient = &fasthttp.Client{ReadTimeout: 5 * time.Second}

// postType，2: json 1: form-data （暂时当0处理） 0：x-www-form-urlencoded
func doHTTP2(method string, urls string, postType int, data map[string]interface{}) string {
	// 初始化请求与响应
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		// 用完需要释放资源
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
		//fasthttpClient.CloseIdleConnections()
	}()
	s := ""
	method = strings.ToUpper(method)
	req.Header.SetMethod(method)
	if method == "POST" {
		// 默认是application/x-www-form-urlencoded
		if postType == 2 {
			req.Header.SetContentType("application/json")
			s = jsontool.MarshalToString(data)
			req.SetBodyString(s)
		} else {
			req.Header.SetContentType("application/x-www-form-urlencoded")
			q := url.Values{}
			for k, v := range data {
				q.Add(k, cast.ToString(v))
			}
			req.SetBodyString(q.Encode())
		}
	}
	if method == "GET" && len(data) > 0 {
		q := url.Values{}
		for k, v := range data {
			q.Add(k, cast.ToString(v))
		}
		if strings.Contains(urls, "?") {
			urls += "&" + q.Encode()
		} else {
			urls += "?" + q.Encode()
		}
	}
	req.SetRequestURI(urls)
	t, _ := time.ParseDuration(conf.GetConfig("system.http.queryTimeOut", "3s").(string))
	err := fasthttpClient.DoTimeout(req, resp, time.Second*t)
	if err != nil {
		logData := make(map[string]string)
		logData["url"] = urls
		logData["error"] = err.Error()
		logData["method"] = method
		logData["postContent"] = s
	}

	b := resp.Body()
	return string(b)
}

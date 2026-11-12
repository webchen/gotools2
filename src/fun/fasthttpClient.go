package fun

import (
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cast"
	"github.com/webchen/gotools2/src/base/conf"

	"github.com/webchen/gotools2/src/base/jsontool"
	"github.com/webchen/gotools2/src/util/logs"

	"github.com/valyala/fasthttp"
)

var fasthttpClient = &fasthttp.Client{ReadTimeout: 5 * time.Second}

// 超时配置缓存：避免每次外部请求都去读配置（读配置要拿 RWMutex + 拆 key）
var httpTimeoutNs atomic.Int64
var httpTimeoutCheckAt atomic.Int64

// 超时缓存有效期，改配置后最长 5 秒生效
const httpTimeoutTTL = 5 * time.Second

// httpQueryTimeout 读取 system.http.queryTimeOut，解析失败或非正数回落 3s
func httpQueryTimeout() time.Duration {
	now := time.Now().UnixNano()
	if now-httpTimeoutCheckAt.Load() < int64(httpTimeoutTTL) {
		if d := httpTimeoutNs.Load(); d > 0 {
			return time.Duration(d)
		}
	}
	raw := strings.TrimSpace(cast.ToString(conf.GetConfig("system.http.queryTimeOut", "3s")))
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		d = 3 * time.Second
	}
	httpTimeoutNs.Store(int64(d))
	httpTimeoutCheckAt.Store(now)
	return d
}

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
	err := fasthttpClient.DoTimeout(req, resp, httpQueryTimeout())
	if err != nil {
		logData := make(map[string]string)
		logData["url"] = urls
		logData["error"] = err.Error()
		logData["method"] = method
		logData["postContent"] = s
		logs.Error("http request fail", logData)
	}

	b := resp.Body()
	return string(b)
}

package fun

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/webchen/gotools2/src/base/conf"
	"github.com/webchen/gotools2/src/base/jsontool"

	"github.com/webchen/gotools2/src/util/logs"

	"github.com/spf13/cast"
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 10 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:        500,
	IdleConnTimeout:     10 * time.Second,
	TLSHandshakeTimeout: 5 * time.Second,
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
}

var httpClient = http.Client{Timeout: 1 * time.Second, Transport: transport}

// 重试次数
var times int = 2

// HTTPGet GET请求
func HTTPGet(url string) string {
	strs := ""
	for j := 1; j <= times; j++ {
		strs = doHTTP("GET", url, 0, nil)
		if strs != "" {
			break
		}
		time.Sleep(time.Millisecond * 20)
	}
	return strs
}

// postType，2: json 1: form-data （暂时当0处理） 0：x-www-form-urlencoded
func doHTTP(method string, url string, postType int, jsonMap map[string]interface{}) string {
	debugBaseGet := cast.ToInt(conf.GetConfig("system.http.debugBaseGet", 0))
	rd := GetReqSeqId()
	if debugBaseGet == 1 {
		logs.Info("doHTTP query [%s] : %s url: %s , data: %#v", rd, method, url, jsonMap)
	}
	s := doHTTP2(method, url, postType, jsonMap)
	if debugBaseGet == 1 {
		logs.Info("doHTTP response [%s] : %s ", rd, s)
	}
	return s
	/*
		strs := ""
		var err error
		req := &http.Request{}
		if method == "GET" {
			req, err = http.NewRequest(method, url, nil)
			if logs.ErrorProcess(err, "unable to create http GET request") {
				return ""
			}
		}
		if method == "POST" {
			strs = jsontool.MarshalToString(jsonMap)
			req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(strs)))
			if logs.ErrorProcess(err, "unable to create http POST request") {
				return ""
			}
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			logs.Query("post: [%s] [%+v], error: %+v", url, strs, err)
		}
		logs.Query("post: [%s] [%+v], success: %s", url, strs, resp.Body)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return string(body)
	*/
}

// HTTPGetSuccess 获取返回正确的请求的值
func HTTPGetSuccess(url string) map[string]interface{} {
	data := HTTPBaseGet(url)
	if it, ok := (data).(map[string]interface{}); ok {
		return it
	}
	return make(map[string]interface{})
}

// HTTPBaseGet 获取数据，返回interface，可逐个类型判断
func HTTPBaseGet(url string) interface{} {
	data := make(map[string]interface{})
	strData := HTTPGet(url)

	if len(strData) == 0 {
		logs.Warning(fmt.Sprintf("http请求 [%s] 返回空", url), "", false)
		return data
	}

	jsontool.LoadFromString(strData, &data)

	if _, ok := data["code"]; !ok {
		logs.Warning(fmt.Sprintf("http请求 [%s] 返回data [%s] 不正确", url, data), "", false)
		return nil
	}

	code := cast.ToInt(data["code"])

	if code != 1 {
		logs.Warning(fmt.Sprintf("http [%s] 请求返回data [%+v] 不正确", url, data), "", false)
		return nil
	}

	if val, ok := data["data"].(interface{}); ok {
		return val
	}
	logs.Warning(fmt.Sprintf("http [%s] 请求 data返回 nil", url), "", false)
	return nil
}

// HTTPGetListSuccess 获取数据列表
func HTTPGetListSuccess(url string) []interface{} {
	data := HTTPBaseGet(url)
	if d, ok := (data).([]interface{}); ok {
		return d
	}
	return nil
}

// HTTPServiceGetSuccess 获取GET数据
func HTTPServiceGetSuccess(url string) map[string]interface{} {
	return HTTPGetSuccess(url)
}

// HTTPPostJSON 发数POST请求
func HTTPPostJSON(url string, jsonMap map[string]interface{}) string {
	strs := ""
	for j := 1; j <= times; j++ {
		strs = doHTTP("POST", url, 2, jsonMap)
		if strs != "" {
			break
		}
	}
	return strs
}

// HTTPServicePostJSON 发送远程POST请求
func HTTPServicePostJSON(url string, jsonMap map[string]interface{}) string {
	return HTTPPostJSON(url, jsonMap)
}

func HTTPServicePostJSONSuccess(url string, jsonMap map[string]interface{}) map[string]interface{} {
	r := HTTPPostJSON(url, jsonMap)
	return getSuccessData(r)
}

func getSuccessData(s string) map[string]interface{} {
	data := make(map[string]interface{})
	jsontool.LoadFromString(s, &data)
	if len(data) == 0 {
		return data
	}

	code := cast.ToInt(data["code"])
	if code != 1 {
		return data
	}

	d, ok := data["data"].(map[string]interface{})
	if ok {
		return d
	}
	return data
}

func getSuccessDataList(s string) []interface{} {
	data := make(map[string]interface{})
	rr := make([]interface{}, 0)
	jsontool.LoadFromString(s, &data)
	if len(data) == 0 {
		return rr
	}

	code := cast.ToInt(data["code"])
	if code != 1 {
		return rr
	}

	d, ok := data["data"].([]interface{})
	if ok {
		return d
	}
	return rr
}

func HTTPPostForm(url string, jsonMap map[string]interface{}) string {
	return doHTTP("POST", url, 0, jsonMap)
}

func HTTPServicePostFormSuccess(url string, jsonMap map[string]interface{}) map[string]interface{} {
	r := HTTPPostForm(url, jsonMap)
	return getSuccessData(r)
}

func HTTPServicePostFormListSuccess(url string, jsonMap map[string]interface{}) []interface{} {
	r := HTTPPostForm(url, jsonMap)
	return getSuccessDataList(r)
}

func HTTPServicePostJsonListSuccess(url string, jsonMap map[string]interface{}) []interface{} {
	r := HTTPPostJSON(url, jsonMap)
	return getSuccessDataList(r)
}

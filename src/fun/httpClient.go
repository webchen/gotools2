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
		time.Sleep(time.Millisecond * 100)
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

// HTTPGetSuccess 获取返回正确的请求的值(data字段的值)，忽略错误
func HTTPGetSuccess(url string) map[string]interface{} {
	data := HTTPGet(url)
	return getSuccessData(data)
}

// HTTPGetListSuccess 获取数据列表，忽略错误
func HTTPGetListSuccess(url string) []interface{} {
	data := HTTPGet(url)
	return getSuccessDataList(data)
}

// HTTPServiceGetSuccess 获取GET数据
func HTTPServiceGetSuccess(url string) map[string]interface{} {
	return HTTPGetSuccess(url)
}

// ------------  post json 相关方法 start -----------
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
func HTTPServicePostJSON(url string, jsonMap map[string]interface{}) (map[string]interface{}, error) {
	r := HTTPPostJSON(url, jsonMap)
	return getBaseData(r)
}

func HTTPServicePostJSONSuccess(url string, jsonMap map[string]interface{}) map[string]interface{} {
	r := HTTPPostJSON(url, jsonMap)
	return getSuccessData(r)
}

func HTTPServicePostJsonList(url string, jsonMap map[string]interface{}) ([]interface{}, error) {
	r := HTTPPostJSON(url, jsonMap)
	return getBaseList(r)
}

func HTTPServicePostJsonListSuccess(url string, jsonMap map[string]interface{}) []interface{} {
	r := HTTPPostJSON(url, jsonMap)
	return getSuccessDataList(r)
}

// ------------  post json 相关方法 end  -----------

func getData(s string) map[string]interface{} {
	data := make(map[string]interface{})
	jsontool.LoadFromString(s, &data)
	return data
}

func getBaseData(s string) (map[string]interface{}, error) {
	data := getData(s)
	code := cast.ToInt(data["code"])
	msg := cast.ToString(data["message"])
	if code != 1 {
		if msg == "" {
			msg = "接口返回状态错误"
		}
		return nil, fmt.Errorf(msg)
	}
	d, ok := data["data"].(map[string]interface{})
	if ok {
		return d, nil
	}
	return nil, fmt.Errorf(msg)
}

func getSuccessData(s string) map[string]interface{} {
	data, err := getBaseData(s)
	if err != nil {
		return nil
	}
	return data
}

// 获取处理好的信息（data字段是第一个返回值。有错误返回nil。）
func GetFixedData(urls string) (map[string]interface{}, error) {
	s := HTTPGet(urls)
	return getBaseData(s)
}

func getBaseList(s string) ([]interface{}, error) {
	data := make(map[string]interface{})
	jsontool.LoadFromString(s, &data)
	if len(data) == 0 {
		return nil, nil
	}

	code := cast.ToInt(data["code"])
	msg := cast.ToString(data["message"])
	if code != 1 {
		if msg == "" {
			msg = "接口返回状态错误"
		}
		return nil, fmt.Errorf(msg)
	}

	d, ok := data["data"].([]interface{})
	if ok {
		return d, nil
	}
	return nil, fmt.Errorf(msg)
}

func getSuccessDataList(s string) []interface{} {
	data, err := getBaseList(s)
	if err != nil {
		return nil
	}
	return data
}

func GetFixedList(urls string) ([]interface{}, error) {
	s := HTTPGet(urls)
	return getBaseList(s)
}

// --------- POST FROM start ----------
func HTTPPostForm(url string, jsonMap map[string]interface{}) string {
	return doHTTP("POST", url, 0, jsonMap)
}

func HTTPServicePostForm(url string, jsonMap map[string]interface{}) (map[string]interface{}, error) {
	r := HTTPPostForm(url, jsonMap)
	return getBaseData(r)
}

func HTTPServicePostFormList(url string, jsonMap map[string]interface{}) ([]interface{}, error) {
	r := HTTPPostForm(url, jsonMap)
	return getBaseList(r)
}

func HTTPServicePostFormSuccess(url string, jsonMap map[string]interface{}) map[string]interface{} {
	r := HTTPPostForm(url, jsonMap)
	return getSuccessData(r)
}

func HTTPServicePostFormListSuccess(url string, jsonMap map[string]interface{}) []interface{} {
	r := HTTPPostForm(url, jsonMap)
	return getSuccessDataList(r)
}

// --------- POST FROM end ----------

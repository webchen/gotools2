package estool

import (
	"strings"

	"github.com/webchen/gotools2/src/base/conf"

	"github.com/webchen/gotools2/src/util/logs"

	"github.com/elastic/go-elasticsearch/v8"
)

var esList = make(map[string]*elasticsearch.Client)

func InitES() error {
	var es *elasticsearch.Client
	serverList := make(map[string]interface{})
	serverList = conf.GetConfig("es", serverList).(map[string]interface{})
	var hostEmpty []interface{}
	for k := range serverList {
		host := conf.GetConfig("es."+k+".host", hostEmpty).([]interface{}) //v["host"].([]interface{})
		var hostList []string
		for _, vv := range host {
			hostList = append(hostList, vv.(string))
		}
		user := conf.GetConfig("es."+k+".user", "").(string)
		password := conf.GetConfig("es."+k+".password", "").(string)
		cfg := elasticsearch.Config{
			Addresses: hostList,
			Username:  user,
			Password:  password,
			// ...
		}
		var err error
		es, err = elasticsearch.NewClient(cfg)
		if logs.ErrorProcess(err, "无法初始化ES") {
			return err
		}
		esList[k] = es
	}
	return nil
}

// GetESClient 获取ES客户端
func GetESClient(key string) *elasticsearch.Client {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	if obj, exists := esList[key]; exists {
		return obj
	}
	return nil
}

/*
// WriteLog 往ES里面写LOG
func WriteLog(level string, message string, v ...interface{}) {
	index := (conf.GetConfig("es.index", "gateway_pub")).(string)
	go (func() {
		data := map[string]interface{}{
			"@timestamp": time.Now().Format(time.RFC3339Nano),
			"level":      level,
			"ip":         nettool.GetLocalFirstIPStr(),
			"message":    message,
			"content":    v,
		}
		body := jsontool.MarshalToString(data)
		req := esapi.IndexRequest{
			Index:   index,
			Body:    bytes.NewReader([]byte(body)),
			Refresh: "true",
		}
		res, err := req.Do(context.Background(), es)
		if err != nil || res == nil {
			log.SetPrefix("ESERROR")
			log.Printf("write log error [%+v] [%+v]", data, err)
			return
		}
		defer res.Body.Close()
		if strings.Contains(res.String(), "error") {
			log.Println(res.String())
		}
	})()
}
*/

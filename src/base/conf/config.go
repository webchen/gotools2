package conf

import (
	"fmt"
	"sync"
	"time"

	"github.com/webchen/gotools2/src/base"
	"github.com/webchen/gotools2/src/base/dirtool"
	"github.com/webchen/gotools2/src/base/jsontool"

	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zouyx/agollo/v4"
	apolloConfig "github.com/zouyx/agollo/v4/env/config"
)

// 全局配置变量
var config = make(map[string]map[string]interface{})

var baseConfigData map[string]map[string]interface{}

var loadTime time.Time = time.Now()

var configLock sync.RWMutex

func init() {
	toInit()
}

func toInit() {
	loadBaseConfig()
	if !base.IsBuild() {
		if checkBaseConfigData() {
			if baseConfigData["configType"]["name"] == nil {
				//log.Println("consul或apollo配置不存在，不更新配置")
			} else {
				configType := strings.TrimSpace(strings.ToLower(baseConfigData["configType"]["name"].(string)))
				if configType == "apollo" {
					initApollo()
				}
				if configType == "consul" {
					initConsulClient()
					initConsul()
				}
			}
		}
	}

	initLocal()
}

func initLocal() {
	configLock.Lock()
	defer configLock.Unlock()
	config = make(map[string]map[string]interface{})
	dir := dirtool.GetConfigPath()
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext == ".json" {
			fileName := strings.ReplaceAll(strings.ReplaceAll(path, dir, ""), ext, "")
			fileName = strings.ReplaceAll(fileName, string(os.PathSeparator), "/")
			conf := make(map[string]interface{})
			jsontool.LoadFromFile(path, &conf)
			config[fileName] = conf
		}
		return nil
	})
	loadTime = time.Now()
}

func Reload() {
	initLocal()
}

func checkBaseConfigData() bool {
	return baseConfigData != nil
}

func loadBaseConfig() {
	if checkBaseConfigData() {
		return
	}
	f := dirtool.GetConfigPath() + "baseConfig.json"
	exists, _ := dirtool.PathExist(f)
	if !exists {
		fmt.Println("tips, " + f + " not exists")
		return
	}
	jsontool.LoadFromFile(f, &baseConfigData)
}

func initConsul() {
	if baseConfigData["consul"] == nil {
		return
	}
	prefix := baseConfigData["consul"]["folder"].(string)
	for _, v := range baseConfigData["consul"]["files"].([]interface{}) {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		r, _, err := consulClient.KV().Get(prefix+v.(string), nil)
		if err != nil {
			log.Println("\n", "config ", v, " read from consul error...")
			continue
		}
		if r == nil {
			log.Println("获取配置[" + v.(string) + "]失败，不更新")
			continue
		}
		fileName := dirtool.GetConfigPath() + v.(string) + ".json"
		os.WriteFile(fileName, r.Value, 0x666)
		fmt.Println("write config file : ", fileName)
	}
}

func initApollo() {
	if baseConfigData == nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			base.LogPanic("initApollo error", p)
		}
	}()

	c := &apolloConfig.AppConfig{
		AppID:         baseConfigData["appolo"]["appID"].(string),
		Cluster:       baseConfigData["appolo"]["cluster"].(string),
		IP:            baseConfigData["apollo"]["host"].(string),
		NamespaceName: baseConfigData["apollo"]["namespace"].(string),
		Secret:        baseConfigData["apollo"]["secret"].(string),
	}
	//	agollo.SetLogger(&log.DefaultLogger{})
	client, _ := agollo.StartWithConfig(func() (*apolloConfig.AppConfig, error) {
		return c, nil
	})

	cache := client.GetConfigCache(c.NamespaceName)
	cache.Range(func(key, value interface{}) bool {
		configFilePath := dirtool.GetConfigPath() + key.(string) + ".json"
		os.WriteFile(configFilePath, []byte(value.(string)), 0777)
		//fmt.Printf("key: %+v   val:%+v\n", key, value)
		return true
	})

	//	value, _ := cache.Get("es")
	//	fmt.Printf("%+v\n%+v\n", cache, value)
}

// GetConfig 获取JSON的配置，key支持"."操作，如：GetConfig("conf.runtime")，表示获取conf.json文件里面，runtime的值
func GetConfig(key string, def interface{}) interface{} {
	configLock.RLock()
	defer configLock.RUnlock()

	defer func() {
		recover()
	}()
	arr := strings.Split(key, ".")
	if len(arr) == 0 {
		return def
	}
	if len(arr) == 1 {
		if config[arr[0]] == nil {
			return def
		}
		return config[arr[0]]
	}
	confDeep := config[arr[0]][arr[1]]
	if len(arr) == 2 {
		if confDeep == nil {
			return def
		}
		return confDeep
	}
	for j := 2; j < len(arr); j++ {
		c, _ := confDeep.(interface{})
		if c == nil {
			return def
		}
		confDeep = confDeep.(map[string]interface{})[arr[j]]
		if confDeep == nil {
			return def
		}
	}
	return confDeep
}

func GetEnv(key, defaultValue string) string {
	if v, ex := os.LookupEnv(key); ex {
		return v
	}
	return defaultValue
}

func GetLoadTime() time.Time {
	return loadTime
}

func SaveConfig(jsonData string, fileName string) error {
	var v interface{}
	err := jsontool.LoadFromString2(jsonData, &v)
	if err != nil {
		return err
	}
	p := dirtool.GetConfigPath() + fileName + ".json"
	err = os.WriteFile(p, []byte(jsonData), 0x666)
	return err
}

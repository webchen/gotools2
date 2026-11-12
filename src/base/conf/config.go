package conf

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cast"
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
var config = make(map[string]map[string]any)

var baseConfigData map[string]map[string]any

var loadTime time.Time = time.Now()

var configLock sync.RWMutex

// 配置读缓存：key -> 解析结果，受 configLock 保护，initLocal 重建配置时整体失效
var configCache = make(map[string]any)

// 配置版本号：initLocal 每次重建配置时自增，用于丢弃并发期间构建的旧缓存
var configVersion atomic.Uint64

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
	config = make(map[string]map[string]any)
	dir := dirtool.GetConfigPath()
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext == ".json" {
			fileName := strings.ReplaceAll(strings.ReplaceAll(path, dir, ""), ext, "")
			fileName = strings.ReplaceAll(fileName, string(os.PathSeparator), "/")
			conf := make(map[string]any)
			jsontool.LoadFromFile(path, &conf)
			config[fileName] = conf
		}
		return nil
	})
	// 配置已整体重建，读缓存全部失效
	configCache = make(map[string]any)
	configVersion.Add(1)
	loadTime = time.Now()
}

func Reload() {
	initLocal()
}

func checkBaseConfigData() bool {
	// 新增判断ENV变量，但是优先从baseConfig.json加载
	if baseConfigData != nil {
		return true
	}
	loadBaseConfigFromEnv()
	return baseConfigData != nil
}

func loadBaseConfigFromEnv() {
	open := cast.ToInt(GetEnv("CONSUL_OPEN", "0"))
	folder := strings.TrimSpace(GetEnv("CONSUL_FOLDER", ""))
	token := strings.TrimSpace(GetEnv("CONSUL_TOKEN", ""))
	host := strings.TrimSpace(GetEnv("CONSUL_HTTP_ADDR", ""))
	files := strings.TrimSpace(GetEnv("CONSUL_CONFIG_FILES", ""))

	ff := make([]any, 0)
	if files != "" {
		for _, v := range strings.Split(files, ",") {
			ff = append(ff, v)
		}
	}

	if open == 1 && folder != "" && host != "" && files != "" && token != "" {
		baseConfigData = make(map[string]map[string]any)
		baseConfigData["configType"] = map[string]any{
			"name": "consul",
		}
		baseConfigData["consul"] = map[string]any{
			"folder": folder,
			"files":  ff,
			"token":  token,
			"server": []any{host},
		}

		f := dirtool.GetConfigPath() + "baseConfig.json"
		os.WriteFile(f, []byte(jsontool.MarshalToString(baseConfigData)), 0666)
	}
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
	for _, v := range baseConfigData["consul"]["files"].([]any) {
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
		err = os.WriteFile(fileName, r.Value, 0666)
		fmt.Println("write config file : ", fileName, " error : ", err)
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
	client, _ := agollo.StartWithConfig(func() (*apolloConfig.AppConfig, error) {
		return c, nil
	})

	cache := client.GetConfigCache(c.NamespaceName)
	cache.Range(func(key, value any) bool {
		configFilePath := dirtool.GetConfigPath() + key.(string) + ".json"
		os.WriteFile(configFilePath, []byte(value.(string)), 0777)
		return true
	})
}

// GetConfig 获取JSON的配置，key支持"."操作，如：GetConfig("conf.runtime")，表示获取conf.json文件里面，runtime的值
// 命中缓存直接返回；未命中才拆分 key 并逐层查询，结果非空则写回缓存（Reload 时缓存整体失效）
func GetConfig(key string, def any) any {
	configLock.RLock()
	if v, ok := configCache[key]; ok {
		configLock.RUnlock()
		return v
	}
	ver := configVersion.Load()
	v, found := resolveConfigLocked(key)
	configLock.RUnlock()

	if found {
		// 写缓存要独占锁；若期间发生了 Reload，本次结果作废，不写入
		configLock.Lock()
		if configVersion.Load() == ver {
			configCache[key] = v
		}
		configLock.Unlock()
		return v
	}
	return def
}

// resolveConfigLocked 在持有 configLock 的前提下逐层解析 key，缺失或类型不符返回 false
func resolveConfigLocked(key string) (any, bool) {
	arr := strings.Split(key, ".")
	if len(arr) == 0 || arr[0] == "" {
		return nil, false
	}
	cur, ok := config[arr[0]]
	if !ok || cur == nil {
		return nil, false
	}
	if len(arr) == 1 {
		return cur, true
	}
	var node any = cur
	for _, part := range arr[1:] {
		m, ok := node.(map[string]any)
		if !ok {
			return nil, false
		}
		node, ok = m[part]
		if !ok || node == nil {
			return nil, false
		}
	}
	return node, true
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
	var v any
	err := jsontool.LoadFromString2(jsonData, &v)
	if err != nil {
		return err
	}
	p := dirtool.GetConfigPath() + fileName + ".json"
	err = os.WriteFile(p, []byte(jsonData), 0x666)
	return err
}

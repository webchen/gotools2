package pool

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gotools2/src/base/conf"

	"gotools2/src/base"

	"gotools2/src/util/logs"

	"github.com/spf13/cast"

	"github.com/go-redis/redis/v8"
)

// Ctx redis的CTX
var Ctx = context.Background()

var clientList = make(map[string]*redis.Client)

func init() {
	if base.IsBuild() {
		return
	}
	var redisList map[string]interface{}
	redisList = conf.GetConfig("redis", redisList).(map[string]interface{})
	if len(redisList) == 0 {
		logs.Warning("redis config is nil", nil, false)
		return
	}

	for k, v := range redisList {
		vv := v.(map[string]interface{})

		host := vv["host"].(string)    // conf.GetConfig("redis."+k+".host", "").(string)
		port := cast.ToInt(vv["port"]) //conf.GetConfig("redis."+k+".port", "").(string)
		db := cast.ToInt(vv["db"])     // conf.GetConfig("redis."+k+".db", "0").(string)
		auth := vv["auth"].(string)    // conf.GetConfig("redis."+k+".auth", "").(string)
		poolSize := cast.ToInt(vv["PoolSize"])
		minIdleConns := cast.ToInt(vv["MinIdleConns"])
		c := redis.NewClient(&redis.Options{
			Addr:         host + ":" + strconv.FormatUint(uint64(port), 10),
			Password:     auth,    // no password set
			DB:           int(db), // use default DB
			PoolSize:     int(poolSize),
			MinIdleConns: int(minIdleConns),
			PoolTimeout:  time.Second * 2,
			IdleTimeout:  time.Second * 2,
		})
		clientList[k] = c
	}
}

// GetClient 获取对象
func GetClient(key string) *redis.Client {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	if v, ok := clientList[key]; ok {
		return v
	}
	return nil
}

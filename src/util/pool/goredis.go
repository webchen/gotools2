package pool

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/webchen/gotools2/src/base/conf"

	"github.com/webchen/gotools2/src/util/logs"

	"github.com/spf13/cast"

	"github.com/go-redis/redis/v8"
)

// Ctx redis的CTX
var Ctx = context.Background()

var redisClientList = make(map[string]*redis.Client)

func InitRedis() error {
	/*
		if base.IsBuild() {
			return nil
		}
	*/
	var redisList map[string]interface{}
	redisList = conf.GetConfig("redis", redisList).(map[string]interface{})
	if len(redisList) == 0 {
		return fmt.Errorf("redis 配置为空")
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
		_, err := c.Ping(Ctx).Result()
		if err != nil {
			logs.Warning("redis 连接失败", err, false)
			return err
		}
		redisClientList[k] = c
	}

	return nil
}

// GetClient 获取对象
func GetRedisClient(key string) *redis.Client {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	if v, ok := redisClientList[key]; ok {
		return v
	}
	logs.Warning("redis client ["+key+"] 不存在", redisClientList, false)
	panic("redis client [" + key + "] 不存在")
}

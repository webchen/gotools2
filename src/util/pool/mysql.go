package pool

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cast"

	"github.com/webchen/gotools2/src/base"
	"github.com/webchen/gotools2/src/base/conf"
	"github.com/webchen/gotools2/src/util/logs"

	"xorm.io/xorm"
	xlog "xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

var mysqlList = make(map[string]*xorm.Engine)

// mysqlList 读写锁：InitMysql 写、GetMysqlClient 读
var mysqlListLock sync.RWMutex

var dbLog *log.Logger
var dbLogOnce sync.Once

// sqlLog sql 日志句柄只创建一次，避免重复开文件与重复起轮转 goroutine
func sqlLog() *log.Logger {
	dbLogOnce.Do(func() {
		dbLog = base.CreateLogFileAccess("sql")
	})
	return dbLog
}

func InitMysql() error {
	mysqlListLock.Lock()
	defer mysqlListLock.Unlock()
	/*
		if base.IsBuild() {
			return nil
		}
	*/
	dbLog := sqlLog()

	list := conf.GetConfig("mysql", nil).(map[string]interface{})
	if len(list) == 0 {
		return fmt.Errorf("mysql 配置为空")
	}
	for k, v := range list {
		vv := v.(map[string]interface{})
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True", vv["user"], vv["password"], vv["host"], vv["port"], vv["db"], vv["charset"])

		db, err := xorm.NewEngine("mysql", dsn)
		if logs.ErrorProcess(err, "connect to mysql fail 1", dsn) {
			return err
		}

		err = db.Ping()
		if logs.ErrorProcess(err, "connect to mysql fail 2", dsn) {
			return err
		}
		db.SetMapper(names.SnakeMapper{})
		//db.ShowSQL(cast.ToBool(vv["showSQL"]))
		db.SetMaxOpenConns(cast.ToInt(vv["maxOpen"]))
		db.SetMaxIdleConns(cast.ToInt(vv["maxIdle"]))
		db.TZLocation, _ = time.LoadLocation("Asia/Shanghai")

		l := &xlog.SimpleLogger{
			DEBUG: dbLog,
			ERR:   dbLog,
			INFO:  dbLog,
			WARN:  dbLog,
		}
		l.ShowSQL(cast.ToBool(vv["showSQL"]))
		dbLog.SetFlags(log.Ldate | log.Lmicroseconds)
		db.SetLogger(l)

		mysqlList[k] = db
	}
	return nil
}

// GetMysqlClient 获取对象
func GetMysqlClient(key string) *xorm.Engine {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	if v, ok := getMysqlClientLocked(key); ok {
		return v
	}
	logs.Warning("mysql client ["+key+"] 不存在", mysqlClientSnapshot(), false)
	panic("mysql client [" + key + "] 不存在")
}

// getMysqlClientLocked 读锁内取引擎
func getMysqlClientLocked(key string) (*xorm.Engine, bool) {
	mysqlListLock.RLock()
	defer mysqlListLock.RUnlock()
	v, ok := mysqlList[key]
	return v, ok
}

// mysqlClientSnapshot 读锁内做一份拷贝，供日志/诊断使用
func mysqlClientSnapshot() map[string]*xorm.Engine {
	mysqlListLock.RLock()
	defer mysqlListLock.RUnlock()
	m := make(map[string]*xorm.Engine, len(mysqlList))
	for k, v := range mysqlList {
		m[k] = v
	}
	return m
}

package pool

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cast"

	"github.com/webchen/gotools2/src/base"
	"github.com/webchen/gotools2/src/base/conf"
	"github.com/webchen/gotools2/src/util/logs"

	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var mysqlList = make(map[string]*xorm.Engine)

//var dbLog *log.Logger

func InitMysql() error {
	if base.IsBuild() {
		return nil
	}
	//dbLog = base.CreateLogFileAccess("db.log")

	list := conf.GetConfig("mysql", nil).(map[string]interface{})

	for k, v := range list {
		vv := v.(map[string]interface{})
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True", vv["user"], vv["password"], vv["host"], vv["port"], vv["db"], vv["charset"])

		db, err := xorm.NewEngine("mysql", dsn)
		if logs.ErrorProcess(err, "connect to mysql fail 1") {
			return err
		}

		err = db.Ping()
		if logs.ErrorProcess(err, "connect to mysql fail 2") {
			return err
		}
		db.SetMapper(names.SnakeMapper{})
		//db.ShowSQL(true)
		db.SetMaxOpenConns(cast.ToInt(vv["maxOpen"]))
		db.SetMaxIdleConns(cast.ToInt(vv["maxIdle"]))
		db.TZLocation, _ = time.LoadLocation("Asia/Shanghai")
		//db.ShowSQL(false)
		//db.SetLogger(dbLog)
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
	if v, ok := mysqlList[key]; ok {
		return v
	}
	return nil
}

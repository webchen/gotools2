package logs

import (
	"fmt"
	"strings"

	"github.com/webchen/gotools2/src/base"
	"github.com/webchen/gotools2/src/base/conf"

	"log"
	"os"
	"time"

	"github.com/spf13/cast"
	//	log "github.com/sirupsen/logrus"
)

// debug -> 0 info/readmq -> 1 warning/query -> 2 error/message -> 3 critial -> 4 exit -> 9

var fileLogger *log.Logger
var cmdLogger *log.Logger

// 日志等级
var cmdLevel int = 0
var fileLevel int = 0

func init() {
	cmdLevel = cast.ToInt(conf.GetConfig("conf.logCmdLevel", 0.0))
	fileLevel = cast.ToInt(conf.GetConfig("conf.logFileLevel", 0.0))
	fileLogger = access("log")
	cmdLogger = newCmdLogger("")
}

// 初始化cmd环境下的logger对象
func newCmdLogger(level string) *log.Logger {
	l := new(log.Logger)
	l.SetPrefix("[" + level + "] ")
	l.SetFlags(log.Lmicroseconds)

	l.SetOutput(os.Stdout)
	return l
}

// Debug 日志
func Debug(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)

	fileLogger.SetPrefix("[Debug] ")
	cmdLogger.SetPrefix("[Debug] ")

	if fileLevel == 0 {
		fileLogger.Println(s)
	}

	if cmdLevel == 0 {
		cmdLogger.Println(s)
	}
}

// Info 日志
func Info(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	t := time.Now().Format(time.DateTime + ".999")
	fileLogger.SetPrefix("[info] [" + t + "] ")
	cmdLogger.SetPrefix("[info] [" + t + "] ")

	if fileLevel <= 1 {
		fileLogger.Println(s)
		fileLogger.SetFlags(0)
	}

	if cmdLevel <= 1 {
		cmdLogger.Println(s)
		cmdLogger.SetFlags(0)
	}
}

// Warning 日志
func Warning(message string, data interface{}, withTrace bool) {
	info := &base.LogObj{}
	info.Message = message
	info.Time = time.Now().Format(time.DateTime + ".999")
	info.Level = "Warning"
	info.Data = data

	s := ""
	if withTrace {
		info.Trace = Trace(message)
		s = fmt.Sprintf("[%s] %s %#v %s", info.Time, info.Message, info.Data, info.Trace)
	} else {
		info.Trace = ""
		s = fmt.Sprintf("[%s] %s %#v", info.Time, info.Message, info.Data)
	}

	if fileLevel <= 2 {
		fileLogger.SetPrefix("[" + info.Level + "] ")
		fileLogger.SetFlags(0)
		fileLogger.Println(s)
	}

	if cmdLevel <= 2 {
		cmdLogger.SetPrefix("[" + info.Level + "] ")
		cmdLogger.SetFlags(0)
		cmdLogger.Println(s)
	}
}

// Error 日志
func Error(message string, data interface{}) {
	info := &base.LogObj{}
	info.Message = message
	info.Time = time.Now().Format(time.DateTime + ".999")
	info.Level = "Error"
	info.Data = data
	info.Trace = Trace(message)

	s := fmt.Sprintf("[%s] %s %#v %s", info.Time, info.Message, info.Data, info.Trace)

	if fileLevel <= 3 {
		fileLogger.SetPrefix("[" + info.Level + "] ")
		fileLogger.SetFlags(0)
		fileLogger.Println(s)
	}

	if cmdLevel <= 3 {
		cmdLogger.SetPrefix("[" + info.Level + "] ")
		cmdLogger.SetFlags(0)
		cmdLogger.Println(s)
	}
}

// ErrorProcess 错误处理
func ErrorProcess(err error, msg string, data interface{}) bool {
	if err != nil {
		msg += "\n" + err.Error()
		Error(msg, data)
		return true
	}
	return false
}

// Show 打印一定会显示的信息（用于系统层面）
func Show(format string, v ...interface{}) {
	//cmdLogger.SetPrefix("[show] ")
	//cmdLogger.Println(fmt.Sprintf(format, v...))
	fmt.Printf(format+"\n", v...)
	//log.Info(fmt.Sprintf(format, v...))
}

func access(fileName string) (l *log.Logger) {
	return base.CreateLogFileAccess(fileName)
}

// Trace 对外TRACE
func Trace(v ...interface{}) string {
	return base.TraceInfo(v)
}

// WebAccess WEB端访问日志
func WebAccess(format string, v ...interface{}) {
	open := conf.GetConfig("conf.openWebAccessLog", false).(bool)
	if !open {
		return
	}
	data := fmt.Sprintf(format, v...)
	if strings.Contains(data, "kube-probe/") || strings.Contains(data, "SLBHealthCheck") {
		return
	}
	fileLogger.SetFlags(log.Ldate | log.Lmicroseconds)
	fileLogger.SetPrefix("[ACCESS] ")
	fileLogger.Println(data)
}

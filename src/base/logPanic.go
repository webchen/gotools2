package base

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	"github.com/webchen/gotools2/src/base/dirtool"
	"github.com/webchen/gotools2/src/base/jsontool"
)

var panicLogger *log.Logger
var separator = "---------------"

// LogObj 日志格式
type LogObj struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Time    string      `json:"time"`
	Level   string      `json:"level"`
	Trace   string      `json:"trace"`
}

func init() {
	panicLogger = CreateLogFileAccess("panic")
}

// CreateLogFileAccess 创建文件日志句柄
func CreateLogFileAccess(fileName string) (l *log.Logger) {
	fullFile := LogDir() + fileName + ".log"
	file, err := os.OpenFile(fullFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatalln("创建日志失败：", err)
	}
	l = new(log.Logger)
	l.SetFlags(log.Lmicroseconds)

	ticker := time.NewTicker(time.Minute * 5)
	lock := &sync.RWMutex{}
	Go(func() {
		for {
			<-ticker.C
			info, _ := os.Stat(fullFile)
			// 500MB
			if info.Size() >= 1024*1024*500 {
				lock.Lock()
				file.Close()
				//os.Remove(fullFile)
				cmdFile := getClearLogCmdPath(fullFile, fileName)
				cmd := exec.Command("sh", cmdFile)
				cmd.Run()
				file, _ := os.OpenFile(fullFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
				l.SetOutput(file)
				lock.Unlock()
			}
		}
	})
	l.SetOutput(file)
	return l
}

func getClearLogCmdPath(fullFile string, f string) (path string) {
	if strings.TrimSpace(fullFile) == "" || strings.TrimSpace(f) == "" || IsWIN() {
		return ""
	}
	info := fmt.Sprintf(
		`
# /usr/bin
echo '' > %s
`, fullFile)
	path = dirtool.GetBasePath() + f + ".sh"
	os.WriteFile(path, []byte(info), 0777)
	return path
}

// LogDir 日志文件夹
func LogDir() string {
	//var dirPath = ""
	/*
		if IsWIN() {
			dirPath, _ = os.Getwd()
			dirPath += string(os.PathSeparator) + "log" + string(os.PathSeparator)
		} else {
			dirPath = "/data/"
		}
	*/
	cfg := make(map[string]interface{})
	jsontool.LoadFromFile(dirtool.GetBasePath()+"config"+string(os.PathSeparator)+"system.json", &cfg)

	dirPath := cast.ToString(cfg["logdir"])
	/*
		if cfg != nil && cfg["logdir"] != nil {
			data, ok := cfg["logdir"].(string)
			if ok {
				dirPath = data
			}
		}
	*/
	if dirPath == "" {
		dirPath, _ = os.Getwd()
		dirPath += string(os.PathSeparator) + "log" + string(os.PathSeparator)
	}
	dirtool.MustCreateDir(dirPath)
	return dirPath
}

// LogPanic PANIC日志
func LogPanic(message string, data interface{}) {
	info := &LogObj{}
	info.Message = message
	info.Time = time.Now().Format(time.DateTime + ".999")
	info.Level = "Panic"
	info.Trace = TraceInfo(message)

	s := fmt.Sprintf("[%s] %s %#v", info.Time, message, info.Data)

	panicLogger.SetPrefix("")
	panicLogger.SetFlags(0)
	panicLogger.Println(s)

	log.SetPrefix("[panic]")
	//log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.Println(s)

}

// LogPanicErr 带ERROR的日志
func LogPanicErr(err interface{}, message string) {
	if err != nil {
		LogPanic(message, err)
	}
	defer func() {
		if r := recover(); r != nil {
			LogPanic("panic recovered ...", err)
		}
	}()
}

// TraceInfo 返回TRACE信息
func TraceInfo(v interface{}) string {
	errstr := fmt.Sprintf("%+v%s%s", v, fmt.Sprintln(), separator) + fmt.Sprintln()
	i := 2
	for {
		pc, file, line, ok := runtime.Caller(i)
		if !ok || i > 40 {
			break
		}
		errstr += fmt.Sprintf("stack: %d [file: %s:%d] [func: %s]\n", i-1, file, line, runtime.FuncForPC(pc).Name())
		i++
	}
	errstr += separator
	return errstr
}

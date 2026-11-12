package base

import (
	"fmt"
	"log"
	"os"
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
	d := LogDir()
	fullFile := d + fileName + ".log"
	maxSize, maxCount := logRotateConfig()
	w, err := newRotateWriter(fullFile, maxSize, maxCount)
	if err != nil {
		log.Fatalln("创建日志失败：", err)
	}
	l = log.New(w, "", log.Lmicroseconds)
	return l
}

// 日志目录与轮转配置缓存：system.json 只需读一次，之后每个日志句柄复用
var logDirOnce sync.Once
var logDirCache string
var logMaxSizeBytes int64
var logMaxCount int

// LogDir 日志文件夹
func LogDir() string {
	logDirOnce.Do(func() {
		cfg := make(map[string]interface{})
		jsontool.LoadFromFile(dirtool.GetBasePath()+"config"+string(os.PathSeparator)+"system.json", &cfg)
		dirPath := cast.ToString(cfg["logdir"])
		if dirPath == "" {
			dirPath, _ = os.Getwd()
			dirPath += string(os.PathSeparator) + "log" + string(os.PathSeparator)
		}
		logDirCache = dirPath

		sizeMB := cast.ToInt64(cfg["logFileSize"])
		if sizeMB <= 0 {
			sizeMB = 500 // 默认 500MB
		}
		logMaxSizeBytes = sizeMB * 1024 * 1024
		logMaxCount = cast.ToInt(cfg["logFileCount"])
	})
	dirtool.MustCreateDir(logDirCache)
	return logDirCache
}

// logRotateConfig 返回单文件上限（字节）与备份保留份数
func logRotateConfig() (int64, int) {
	LogDir() // 确保配置已加载
	return logMaxSizeBytes, logMaxCount
}

// LogPanic PANIC日志
func LogPanic(message string, data interface{}) {
	info := &LogObj{}
	info.Message = message
	info.Time = time.Now().Format(time.DateTime + ".999")
	info.Level = "Panic"
	info.Data = data
	info.Trace = TraceInfo(message)

	s := fmt.Sprintf("[%s] %s %#v\n%s", info.Time, message, info.Data, info.Trace)

	panicLogger.SetPrefix("")
	panicLogger.SetFlags(0)
	//panicLogger.SetPrefix("[panic]")
	//log.SetFlags(log.Ldate | log.Lmicroseconds)
	panicLogger.Println(s)
	fmt.Println(s)
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

// maxTraceDepth 堆栈采集深度上限（runtime.Caller 的 i 终止值，最多 maxTraceDepth-1 帧）
// 默认 10，可由 logs 包按 conf.logTraceDepth 覆盖；上限 40 防止采过深拖慢日志
var maxTraceDepth = 10

// SetTraceDepth 设置堆栈采集深度上限；取值超出 1..40 时忽略，保持当前值
func SetTraceDepth(n int) {
	if n > 0 && n <= 40 {
		maxTraceDepth = n
	}
}

// TraceInfo 返回TRACE信息
func TraceInfo(v interface{}) string {
	// 用 Builder 一次成型：原实现 errstr += fmt.Sprintf(...) 每帧整串重拷，39 帧时是 O(n²)
	var sb strings.Builder
	sb.Grow(512)
	fmt.Fprintf(&sb, "%+v\n%s\n", v, separator)
	for i := 2; i <= maxTraceDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		sb.WriteString("stack: ")
		sb.WriteString(cast.ToString(i - 1))
		sb.WriteString(" [file: ")
		sb.WriteString(file)
		sb.WriteString(":")
		sb.WriteString(cast.ToString(line))
		sb.WriteString("] [func: ")
		sb.WriteString(runtime.FuncForPC(pc).Name())
		sb.WriteString("]\n")
	}
	sb.WriteString(separator)
	return sb.String()
}

package base

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	file, err := os.OpenFile(fullFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatalln("创建日志失败：", err)
	}
	l = new(log.Logger)
	l.SetFlags(log.Lmicroseconds)

	ticker := time.NewTicker(time.Minute * 1)
	lock := &sync.RWMutex{}

	cfg := make(map[string]interface{})
	jsontool.LoadFromFile(dirtool.GetBasePath()+"config"+string(os.PathSeparator)+"system.json", &cfg)

	logFileSize := cast.ToInt64(cfg["logFileSize"])
	logFileCount := cast.ToInt(cfg["logFileCount"])

	if logFileSize <= 0 {
		logFileSize = 500 // 默认500MB
	}

	Go(func() {
		for {
			<-ticker.C
			info, _ := os.Stat(fullFile)
			size := 1024 * 1024 * logFileSize
			if info.Size() >= size {
				lock.Lock()
				file.Close()
				cmdFile := getClearLogCmdPath(fullFile, logFileCount, size)
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

func getClearLogCmdPath(fullFile string, total int, size int64) (path string) {
	if strings.TrimSpace(fullFile) == "" || IsWIN() {
		return ""
	}
	fileDir := filepath.Dir(fullFile)
	fileName := filepath.Base(fullFile)
	fileExt := filepath.Ext(fullFile) // 带"."
	fileName = strings.ReplaceAll(fileName, fileExt, "")
	info := `# /user/bin
d=$(date +"%Y%m%d%H%M%S")
fileName="@fileName@"
dir="@fileDir@"
back_dir=${dir}/back
ext="@fileExt@"

n=$(find ${back_dir} -type f -printf '%T+ %p\n' | grep "@fileName@_" | grep "\@fileExt@" | wc -l)
((n2=$n-@total@))
if [ $n2 -gt 0 ]; then
find ${back_dir} -type f -printf '%T+ %p\n' | grep "@fileName@_" | grep "\@fileExt@" | sort | head -n $n2 | xargs rm -rf
fi

logFile=${dir}/${fileName}${ext}
sz=$(ls -l ${logFile} | awk '{print $5}')
if [ $sz -lt @size@ ]; then
	exit 0
else
	mv ${logFile} ${back_dir}/${fileName}_${d}${ext}
	touch ${logFile}
	chmod 0777 ${logFile}
fi

`
	info = strings.ReplaceAll(info, "@fileExt@", fileExt)
	info = strings.ReplaceAll(info, "@fileName@", fileName)
	info = strings.ReplaceAll(info, "@fileDir@", fileDir)
	info = strings.ReplaceAll(info, "@total@", cast.ToString(total))
	info = strings.ReplaceAll(info, "@size@", cast.ToString(size))
	path = dirtool.GetBasePath() + fileName + ".sh"
	os.WriteFile(path, []byte(info), 0777)
	return path
}

// LogDir 日志文件夹
func LogDir() string {
	cfg := make(map[string]interface{})
	jsontool.LoadFromFile(dirtool.GetBasePath()+"config"+string(os.PathSeparator)+"system.json", &cfg)
	dirPath := cast.ToString(cfg["logdir"])
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

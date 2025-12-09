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

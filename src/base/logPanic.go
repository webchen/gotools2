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
			if info.Size() >= 1024*1024*logFileSize {
				/*
					lock.Lock()
					newFile := d + fileName + "_" + time.Now().Format("20060102150405") + ".log"
					newFileObj, err := os.OpenFile(newFile, os.O_WRONLY|os.O_CREATE, 0777)

					fmt.Println(newFile, err)

					i, err := io.Copy(newFileObj, file)
					fmt.Println(i, err)

					newFileObj.Close()
					file.Close()

					if logFileCount > 0 {
						ll := getLogFileList(d)
						count := len(ll)
						if count >= logFileCount {
							// 找出最老的，删了
							m := make([]int, 0)
							for _, v := range ll {
								m = append(m, cast.ToInt(v))
							}
							sort.Ints(m) // 小到大，第1个就是最老的
							os.Remove(d + ll[cast.ToInt64(m[0])])
						}
					}

					//file.Close()
					os.Remove(fullFile)
				*/
				lock.Lock()
				file.Close()
				cmdFile := getClearLogCmdPath(fullFile, logFileCount)
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

/*
func getLogFileList(d string) map[int64]string {
	infos, err := os.ReadDir(d)
	if err != nil {
		return nil
	}
	r := make(map[int64]string)
	for _, v := range infos {
		if !v.IsDir() {
			vv, _ := v.Info()
			r[vv.ModTime().Unix()] = v.Name()
		}
	}
	return r
}
*/

func getClearLogCmdPath(fullFile string, total int) (path string) {
	if strings.TrimSpace(fullFile) == "" || IsWIN() {
		return ""
	}
	fileDir := filepath.Dir(fullFile)
	fileName := filepath.Base(fullFile)
	fileExt := filepath.Ext(fullFile)
	fileName = strings.ReplaceAll(fileName, fileExt, "")
	info := `# /user/bin
d=$(date +"%Y%m%d%H%M%S")
fileName="@fileName@"
dir="@fileDir@"
back_dir=${dir}/back
n=$(find ${back_dir} -type f | wc -l)
if [ $n -gt @total@ ]; then
find ${back_dir} -type f -printf '%T+ %p\n' | sort | head -n 1 | xargs rm -rf
fi
mv ${dir}/${fileName}.log ${back_dir}/${fileName}_${d}.log
chmod 0777 ${dir}/${fileName}.log
echo '' > ${dir}/${fileName}.log`
	info = strings.ReplaceAll(info, "@fileName@", fileName)
	info = strings.ReplaceAll(info, "@dileDir@", fileDir)
	info = strings.ReplaceAll(info, "@total@", cast.ToString(total))
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

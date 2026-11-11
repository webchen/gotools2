package logs

import (
	"fmt"
	"strings"
	"sync/atomic"

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

// WebAccess 开关缓存：避免每个请求都去读配置（读配置要拿 RWMutex + 拆 key）
var webAccessOpen atomic.Bool
var webAccessCheckAt atomic.Int64

// 开关缓存有效期，改配置后最长 5 秒生效
const webAccessCheckTTL = 5 * time.Second

func init() {
	cmdLevel = cast.ToInt(conf.GetConfig("conf.logCmdLevel", 0.0))
	fileLevel = cast.ToInt(conf.GetConfig("conf.logFileLevel", 0.0))
	// 堆栈深度：默认 10 帧，可在 conf 里用 logTraceDepth 配回 40
	base.SetTraceDepth(cast.ToInt(conf.GetConfig("conf.logTraceDepth", 10)))
	fileLogger = access("log")
	// prefix/flags 只在初始化时设置一次，运行期不再改动：
	// 级别与时间由各日志函数自己拼进消息体，靠单次 Println 保证整行原子写出，
	// 避免并发时"设置前缀"和"写行"两步被其它 goroutine 插入，造成时间戳与内容错配。
	fileLogger.SetPrefix("")
	fileLogger.SetFlags(0)
	cmdLogger = newCmdLogger("")
}

// 初始化cmd环境下的logger对象
func newCmdLogger(level string) *log.Logger {
	if level != "" {
		level = "[" + level + "] "
	}
	return log.New(os.Stdout, level, 0)
}

// Debug 日志
func Debug(format string, v ...interface{}) {
	// 文件与命令行都被级别过滤时，直接返回，不做格式化与取时间
	if fileLevel > 0 && cmdLevel > 0 {
		return
	}

	s := "[info] [" + time.Now().Format(time.DateTime+".999") + "] " + fmt.Sprintf(format, v...)

	if fileLevel == 0 {
		fileLogger.Println(s)
	}

	if cmdLevel == 0 {
		cmdLogger.Println(s)
	}
}

// Info 日志
func Info(format string, v ...interface{}) {
	if fileLevel > 1 && cmdLevel > 1 {
		return
	}

	s := "[info] [" + time.Now().Format(time.DateTime+".999") + "] " + fmt.Sprintf(format, v...)

	if fileLevel <= 1 {
		fileLogger.Println(s)
	}

	if cmdLevel <= 1 {
		cmdLogger.Println(s)
	}
}

// Warning 日志
func Warning(message string, data interface{}, withTrace bool) {
	// 级别不够时不做任何拼接，更不做堆栈采集
	if fileLevel > 2 && cmdLevel > 2 {
		return
	}

	var info base.LogObj
	info.Message = message
	info.Time = time.Now().Format(time.DateTime + ".999")
	info.Level = "Warning"
	info.Data = data

	if withTrace {
		info.Trace = Trace(message)
	}

	s := buildLogLine(&info)

	if fileLevel <= 2 {
		fileLogger.Println(s)
	}

	if cmdLevel <= 2 {
		cmdLogger.Println(s)
	}
}

// Error 日志
func Error(message string, data interface{}) {
	// 级别不够时不做堆栈采集（Trace 要遍历调用栈，开销大）
	if fileLevel > 3 && cmdLevel > 3 {
		return
	}

	var info base.LogObj
	info.Message = message
	info.Time = time.Now().Format(time.DateTime + ".999")
	info.Level = "Error"
	info.Data = data
	info.Trace = Trace(message)

	s := buildLogLine(&info)

	if fileLevel <= 3 {
		fileLogger.Println(s)
	}

	if cmdLevel <= 3 {
		cmdLogger.Println(s)
	}
}

// dataMaxItems 集合类数据超过该元素数时只打类型与长度，避免 %#v 递归展开刷屏、拖慢日志
const dataMaxItems = 32

// formatData 格式化日志附加数据：小数据保持 %#v（与改前逐字节一致），超大集合降级为类型+长度
func formatData(data interface{}) string {
	switch x := data.(type) {
	case nil:
		return "<nil>"
	case map[string]interface{}:
		if len(x) > dataMaxItems {
			return fmt.Sprintf("map[string]interface {}(len=%d)", len(x))
		}
	case []interface{}:
		if len(x) > dataMaxItems {
			return fmt.Sprintf("[]interface {}(len=%d)", len(x))
		}
	}
	return fmt.Sprintf("%#v", data)
}

// buildLogLine 单次组装 "[Level] [time] message data [trace]"。
// 与改前「Sprintf("[%s] %s %#v %s") 后再拼 "[Level] "」的输出逐字节一致，但少两次整串拷贝。
func buildLogLine(info *base.LogObj) string {
	var sb strings.Builder
	sb.Grow(256)
	sb.WriteString("[")
	sb.WriteString(info.Level)
	sb.WriteString("] [")
	sb.WriteString(info.Time)
	sb.WriteString("] ")
	sb.WriteString(info.Message)
	sb.WriteString(" ")
	sb.WriteString(formatData(info.Data))
	if info.Trace != "" {
		sb.WriteString(" ")
		sb.WriteString(info.Trace)
	}
	return sb.String()
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

// webAccessEnabled 带 TTL 的访问日志开关，热更新仍然生效（最长延迟 webAccessCheckTTL）
func webAccessEnabled() bool {
	now := time.Now().UnixNano()
	if now-webAccessCheckAt.Load() > int64(webAccessCheckTTL) {
		// 配置值可能是 bool，也可能是字符串 "true"/"1"，统一用 cast 转换，避免类型断言 panic
		webAccessOpen.Store(cast.ToBool(conf.GetConfig("conf.openWebAccessLog", false)))
		webAccessCheckAt.Store(now)
	}
	return webAccessOpen.Load()
}

// WebAccessEnabled 供中间件预判是否值得采集请求体与响应体，关闭时省掉这部分开销
func WebAccessEnabled() bool {
	return webAccessEnabled()
}

// WebAccess WEB端访问日志
func WebAccess(format string, v ...interface{}) {
	WebAccessLine(fmt.Sprintf(format, v...))
}

// WebAccessLine 写入已组装好的访问日志内容（不含 [ACCESS] 前缀与时间）。
// 供调用方用 strings.Builder 拼大报文，省掉 %s 对大字段的整段字符串拷贝。
func WebAccessLine(data string) {
	if !webAccessEnabled() {
		return
	}

	if strings.Contains(data, "kube-probe/") || strings.Contains(data, "SLBHealthCheck") {
		return
	}

	// 原来靠 log 的 Ldate|Lmicroseconds 打日期，现改为自拼，格式与 log 包输出完全一致，
	// 运行期不再 SetFlags/SetPrefix，保证并发安全。
	fileLogger.Println(webAccessLine(data, time.Now()))
}

// webAccessLine 组装访问日志行；时间布局等价于 log 包的 Ldate|Lmicroseconds
func webAccessLine(data string, t time.Time) string {
	return "[ACCESS] " + t.Format("2006/01/02 15:04:05.000000") + " " + data
}

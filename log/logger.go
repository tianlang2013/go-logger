package log

import (
	"sync/atomic"
	"os"
	"sync"
	"io"
	"fmt"
	"time"
	"runtime"
)

const (
	Ldate         = 1 << iota     // 日期格式: 2009/01/23
	Ltime                         // 时间格式 e: 01:23:23
	Lmicroseconds                 // 显示毫秒: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // 显示完整文件路径: /a/b/c/d.go:23
	Lshortfile                    // 显示文件名: d.go:23. overrides Llongfile
	LUTC                          // 以UTC格式输出
	LstdFlags     = Ldate | Ltime // initial values for the standard Log
)
const (
	LvlCrit int32 = iota
	LvlError
	LvlWarn
	LvlInfo
	LvlDebug
	LvlTrace
)

var LLevel int32 = LvlInfo

type Logger struct {
	mu     sync.Mutex //  同步锁
	model  string     // 模块，默认为BASE
	prefix string     // 日志分类，debug info warn trace crit error
	flag   int        // 属性
	out    io.Writer  // 输出器
	buf    []byte     // 内容缓存
}

func New(model string) *Logger {
	return &Logger{out: os.Stderr, flag: LstdFlags | Lshortfile , model: fmt.Sprintf("%s\t", model)}
}

func (self *Logger) SetOutput(w io.Writer) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.out = w
}


func itoa(buf *[]byte, i int, wid int) {
	// 以相反的顺序组合十进制。
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

//格式化标准头信息
func (self *Logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {

	*buf = append(*buf, self.prefix...)
	if self.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if self.flag&LUTC != 0 {
			t = t.UTC()
		}
		if self.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if self.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if self.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	*buf = append(*buf, self.model...)
	if self.flag&(Lshortfile|Llongfile) != 0 {
		if self.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

//格式化输出
func (self *Logger) Output(calldepth int, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		self.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		self.mu.Lock()
	}
	self.buf = self.buf[:0]
	self.formatHeader(&self.buf, now, file, line)
	self.buf = append(self.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		self.buf = append(self.buf, '\n')
	}
	_, err := self.out.Write(self.buf)
	return err
}

func (self *Logger) Trace(msg string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlTrace {
		return
	}

	self.defaultFormat("trace", msg, v...)
}

func (self *Logger) Debug(msg string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlDebug {
		return
	}

	self.defaultFormat("debug", msg, v...)
}

func (self *Logger) Info(msg string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlInfo {
		return
	}

	self.defaultFormat("info", msg, v...)
}

func (self *Logger) Warn(msg string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlWarn {
		return
	}

	self.defaultFormat("warn", msg, v...)
}

func (self *Logger) Error(msg string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlError {
		return
	}

	self.defaultFormat("error", msg, v...)
}

func (self *Logger) Crit(msg string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit {
		return
	}
	self.defaultFormat("crit", msg, v...)
	os.Exit(1)
}

/**
增加格式化支持
*/

func (self *Logger) TraceF(format string, v ...interface{}) {
	self.formatF("trace", format, v...)
}

func (self *Logger) DebugF(format string, v ...interface{}) {
	self.formatF("debug", format, v...)
}

func (self *Logger) InfoF(format string, v ...interface{}) {
	self.formatF("info", format, v...)
}

func (self *Logger) WarnF(format string, v ...interface{}) {
	self.formatF("warn", format, v...)
}

func (self *Logger) ErrorF(format string, v ...interface{}) {
	self.formatF("error", format, v...)
}

func (self *Logger) CritF(format string, v ...interface{}) {
	self.formatF("crit", format, v...)
	os.Exit(1)
}

func (self *Logger) formatF(tag string, format string, v ...interface{}) {
	//defer func() {
	//	if p := recover() ;p !=nil {
	//		self.Output( 3 ,fmt.Sprint( i18.I18_print.GetValue("日志数据打印异常"), format , v ) )
	//	}
	//}()
	self.SetPrefix(tag + "\t")
	self.Output(3, fmt.Sprintf(format, v...))
}

func (self *Logger) defaultFormat(tag string, format string, v ...interface{}) {
	//defer func() {
	//	if p := recover() ;p !=nil {
	//		self.Output( 3 ,fmt.Sprint( i18.I18_print.GetValue("日志数据打印异常"), format , v ) )
	//	}
	//}()

	self.SetPrefix(tag + "\t")
	if len(v) > 0 {
		self.Output(3, fmt.Sprintln(format, v))
	} else {
		self.Output(3, fmt.Sprintln(format))
	}
}

func (self *Logger) SetLevel(level int32) {
	if level < LvlError || level > LvlTrace {
		return
	}
	atomic.StoreInt32(&LLevel, level)
}

// 与crit类似
func (self *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v)
	self.Output(2, s)
	panic(s)
}

// 返回当前的输出属性
func (self *Logger) Flags() int {
	self.mu.Lock()
	defer self.mu.Unlock()
	return self.flag
}

//
func (self *Logger) SetFlags(flag int) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.flag = flag
}

// 返回日志的输出类型，每个类型是固定的，目前没有太太意义
func (self *Logger) Prefix() string {
	self.mu.Lock()
	defer self.mu.Unlock()
	return self.prefix
}

// 配置日志输出的类型
func (self *Logger) SetPrefix(prefix string) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.prefix = prefix
}

// 配置输出接口
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

// 配置属性，用来控制 日期、时间、文件名等输出
func SetFlags(flag int) {
	std.SetFlags(flag)
}

func Trace(msg string, v ...interface{}) {
	std.defaultFormat("trace", msg, v...)
}

func Debug(msg string, v ...interface{}) {
	std.defaultFormat("debug", msg, v...)
}

func Info(msg string, v ...interface{}) {
	std.defaultFormat("info", msg, v...)
}

func Warn(msg string, v ...interface{}) {
	std.defaultFormat("warn", msg, v...)
}

func Error(msg string, v ...interface{}) {
	std.defaultFormat("error", msg, v...)
}

//输出信息并终止整个系统运行
func Crit(msg string, v ...interface{}) {
	std.defaultFormat("crite", msg, v...)
}

/**
增加默认格式化加支持
*/

func TraceF(format string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlTrace {
		return
	}
	std.formatF("trace", format, v...)
}

func DebugF(format string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit {
		return
	}
	std.formatF("debug", format, v...)
}

func InfoF(format string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit {
		return
	}
	std.formatF("info", format, v...)
}

func ErrorF(format string, v ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit {
		return
	}
	std.formatF("error", format, v...)
}

func WarnF(format string, v ...interface{}) {

	std.formatF("warn", format, v...)
}

func CritF(format string, v ...interface{}) {

	std.formatF("crit", format, v...)
	os.Exit(1)
}

//设置日志级别
func SetLevel(level string) {
	switch level {
	case "error":
		std.SetLevel(LvlError)
	case "crit":
		std.SetLevel(LvlCrit)
	case "warn":
		std.SetLevel(LvlWarn)
	case "info":
		std.SetLevel(LvlInfo)
	case "debug":
		std.SetLevel(LvlDebug)
	case "trace":
		std.SetLevel(LvlTrace)
	default:
		std.SetLevel(LvlInfo)
	}
}

func GetLevel() string {
	level := atomic.LoadInt32(&LLevel)
	switch level {
	case LvlError:
		return "error"
	case LvlCrit:
		return "crit"
	case LvlWarn:
		return "warn"
	case LvlInfo:
		return "info"
	case LvlDebug:
		return "debug"
	case LvlTrace:
		return "trace"
	default:
		return "info"
	}
	return ""
}

func SetLevelNum(level int32) {
	std.SetLevel(level)
}

//单独封装输出接口
func Output(calldepth int, s string) error {
	return std.Output(calldepth+1, s) // +1 for this frame.
}

var std = New("base")
 

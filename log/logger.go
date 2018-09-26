package log 

import (
	"fmt"
	"os"
	"sync"
	"io"
	"time"
	"runtime"
	"log"
	"sync/atomic"
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
	LvlCrit int32= iota
	LvlError
	LvlWarn
	LvlInfo
	LvlDebug
	LvlTrace
)


var LLevel int32	= LvlDebug

type Log struct {
	mu     sync.Mutex 	//  同步锁
	model  string		// 模块，默认为BASE
	prefix string     	// 日志分类，debug info warn trace crit error
	flag   int        	// 属性
	out    io.Writer  	// 输出器
	buf    []byte     	// 内容缓存
}

func New(  model string ) *Log {
	return &Log{out: os.Stderr,  flag: Lshortfile|log.LstdFlags , model : fmt.Sprintf("%s\t" ,model ) }
}


func (l *Log) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

var std = New( "base")


func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
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
func (l *Log) formatHeader(buf *[]byte, t time.Time, file string, line int) {

	*buf = append(*buf, l.prefix...)
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	*buf = append(*buf, l.model...)
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
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
func (l *Log) Output(calldepth int, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}



func (l *Log) Trace(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlTrace{
		return
	}
	l.SetPrefix("trace	")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Log) Debug(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlDebug{
		return
	}
	l.SetPrefix("debug	")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Log) Info(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlInfo{
		return
	}
	l.SetPrefix("info	")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Log) Warn(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlWarn{
		return
	}
	l.SetPrefix("warn	")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Log) Error(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlError{
		return
	}
	l.SetPrefix("error	")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Log) Crit(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit{
		return
	}
	l.SetPrefix("crit	")
	l.Output(2, fmt.Sprintln(msg,ctx) )
	os.Exit(1)
}

/**
	增加格式化支持
 */

func (l *Log) TraceF(format string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlTrace{
		return
	}
	l.SetPrefix("trace	")
	l.Output(2, fmt.Sprintf(format,ctx) )
}

func (l *Log) DebugF(format string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlDebug{
		return
	}
	l.SetPrefix("debug	")
	l.Output(2, fmt.Sprintf(format,ctx) )
}

func (l *Log) InfoF(format string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlInfo{
		return
	}
	l.SetPrefix("info	")
	l.Output(2, fmt.Sprintf(format,ctx) )
}

func (l *Log) WarnF(format string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlWarn{
		return
	}
	l.SetPrefix("warn	")
	l.Output(2, fmt.Sprintf(format,ctx) )
}

func (l *Log) ErrorF(format string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlError{
		return
	}
	l.Output(2, fmt.Sprintf(format,ctx) )
}

func (l *Log) CritF(format string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit{
		return
	}
	l.SetPrefix("crit	")
	l.Output(2, fmt.Sprintf(format,ctx) )
	os.Exit(1)
}



func (l *Log) SetLevel(level int32)  {
	if level < LvlError || level > LvlTrace {
		return
	}
	atomic.StoreInt32(&LLevel , level)
}

// 与crit类似
func (l *Log) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Output(2, s)
	panic(s)
}

// 返回当前的输出属性
func (l *Log) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

//
func (l *Log) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

// 返回日志的输出类型，每个类型是固定的，目前没有太太意义
func (l *Log) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

// 配置日志输出的类型
func (l *Log) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
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


func  Trace(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlTrace{
		return
	}
	std.SetPrefix("trace	")
	std.Output(2, fmt.Sprintln(msg,ctx) )
}

func Debug(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlDebug{
		return
	}
	std.SetPrefix("debug	")
	std.Output(2, fmt.Sprintln(msg,ctx) )
}

func  Info(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlInfo{
		return
	}
	std.SetPrefix("info	")
	std.Output(2, fmt.Sprintln(msg,ctx) )
}

func  Warn(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlWarn{
		return
	}
	std.SetPrefix("warn	")
	std.Output(2, fmt.Sprintln(msg,ctx) )
}

func  Error(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlError{
		return
	}
	std.SetPrefix("error	")
	std.Output(2, fmt.Sprintln(msg,ctx) )
}

//输出信息并终止整个系统运行
func  Crit(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit{
		return
	}
	std.SetPrefix("crit	")
	std.Output(2, fmt.Sprintln(msg,ctx) )
	os.Exit(1)
}

/**
	增加默认格式化加支持
 */

func TraceF(format string , v ...interface{})  {
	Trace( fmt.Sprintf(format , v))
}

func DebugF(format string , v ...interface{})  {
	Debug( fmt.Sprintf(format , v))
}

func InfoF(format string , v ...interface{})  {
	Info( fmt.Sprintf(format , v))
}

func ErrorF(format string , v ...interface{})  {
	Error( fmt.Sprintf(format , v))
}

func WarnF(format string , v ...interface{})  {
	Warn( fmt.Sprintf(format , v))
}

func CritF(format string , v ...interface{})  {
	Crit( fmt.Sprintf(format , v))
}

//设置日志级别
func SetLevel(level int32)  {
	std.SetLevel(level)
}

//单独封装输出接口
func Output(calldepth int, s string) error {
	return std.Output(calldepth+1, s) // +1 for this frame.
}


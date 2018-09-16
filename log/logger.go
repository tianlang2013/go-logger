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
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)
const (
	LvlCrit int32= iota
	LvlError
	LvlWarn
	LvlInfo
	LvlDebug
	LvlTrace
)


var LLevel int32	= LvlInfo

type Logger struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	prefix string     // prefix to write at beginning of each line
	flag   int        // properties
	out    io.Writer  // destination for output
	buf    []byte     // for accumulating text to write
}

func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{out: out, prefix: prefix , flag: flag}
}


func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

var std = New(os.Stderr, "", LstdFlags | log.Lshortfile)


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


func (l *Logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
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


func (l *Logger) Output(calldepth int, s string) error {
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



func (l *Logger) Trace(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlTrace{
		return
	}
	l.SetPrefix("trace")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Logger) Debug(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlDebug{
		return
	}
	l.SetPrefix("debug")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Logger) Info(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlInfo{
		return
	}
	l.SetPrefix("info")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Logger) Warn(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlWarn{
		return
	}
	l.SetPrefix("warn")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Logger) Error(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlError{
		return
	}
	l.SetPrefix("error")
	l.Output(2, fmt.Sprintln(msg,ctx) )
}

func (l *Logger) Crit(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit{
		return
	}
	l.SetPrefix("crit")
	l.Output(2, fmt.Sprintln(msg,ctx) )
	os.Exit(1)
}

func (l *Logger) SetLevel(level int32)  {
	if level < LvlError || level < LvlTrace {
		return
	}
	atomic.StoreInt32(&LLevel , level)
}

// Panicln is equivalent to l.Println() followed by a call to panic().
func (l *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Output(2, s)
	panic(s)
}

// Flags returns the output flags for the logger.
func (l *Logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

// SetFlags sets the output flags for the logger.
func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

// Prefix returns the output prefix for the logger.
func (l *Logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

// SetPrefix sets the output prefix for the logger.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}


// SetFlags sets the output flags for the standard logger.
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

func  Crit(msg string, ctx ...interface{}) {
	if atomic.LoadInt32(&LLevel) < LvlCrit{
		return
	}
	std.SetPrefix("crit	")
	std.Output(2, fmt.Sprintln(msg,ctx) )
	os.Exit(1)
}

func SetLevel(level int32)  {
	std.SetLevel(level)
}
func Output(calldepth int, s string) error {
	return std.Output(calldepth+1, s) // +1 for this frame.
}

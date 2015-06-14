package log

import (
	"cham/cham"
	"fmt"
	"runtime"
	"strconv"
	"time"
)

const (
	LDEBUG = iota
	LINFO
	LERROR
)

const (
	LTIME = 1 << iota
	LFILE
	LLEVEL
	LDEFAULT = LTIME | LLEVEL
	LALL     = LTIME | LFILE | LLEVEL
)

const TimeFormat = "2006/01/02 15:04:05"

var Names = [3]string{"Debug", "Info", "Error"}
var logs = cham.NewService("log", Start)

type Logger struct {
	level int
	flag  int
	out   string
}

//if out empty, use os.Stdout
func New(out string, flag int, level int) *Logger {
	cham.Main.Call(logs, cham.PTYPE_GO, OPEN, out)
	return &Logger{level, flag, out}
}

var std = New("", LDEFAULT, LINFO)

func (l *Logger) Output(calldepth int, level int, s string) {
	if l.level > level {
		return
	}
	buf := make([]byte, 0, len(s)+28)
	if l.flag&LTIME != 0 {
		now := time.Now().Format(TimeFormat)
		buf = append(buf, now...)
	}
	if l.flag&LFILE != 0 {
		if _, file, line, ok := runtime.Caller(calldepth); ok {
			buf = append(buf, file...)
			buf = append(buf, ':')
			buf = strconv.AppendInt(buf, int64(line), 10)
		}
	}
	if l.flag&LLEVEL != 0 {
		buf = append(buf, " ["...)
		buf = append(buf, Names[level]...)
		buf = append(buf, "] "...)
	}
	buf = append(buf, s...)
	cham.Main.Notify(logs, cham.PTYPE_GO, WRITE, buf)
}

func (l *Logger) Debug(v ...interface{}) {
	l.Output(2, LDEBUG, fmt.Sprint(v...))
}

func (l *Logger) Debugln(v ...interface{}) {
	l.Output(2, LDEBUG, fmt.Sprintln(v...))
}

func (l *Logger) Debugf(f string, v ...interface{}) {
	l.Output(2, LDEBUG, fmt.Sprintf(f, v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.Output(2, LINFO, fmt.Sprint(v...))
}

func (l *Logger) Infoln(v ...interface{}) {
	l.Output(2, LINFO, fmt.Sprintln(v...))
}

func (l *Logger) Infof(f string, v ...interface{}) {
	l.Output(2, LINFO, fmt.Sprintf(f, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.Output(2, LERROR, fmt.Sprint(v...))
}

func (l *Logger) Errorln(v ...interface{}) {
	l.Output(2, LERROR, fmt.Sprintln(v...))
}

func (l *Logger) Errorf(f string, v ...interface{}) {
	l.Output(2, LERROR, fmt.Sprintf(f, v...))
}

// for std
func Debug(v ...interface{}) {
	std.Output(2, LDEBUG, fmt.Sprint(v...))
}

func Debugln(v ...interface{}) {
	std.Output(2, LDEBUG, fmt.Sprintln(v...))
}

func Debugf(f string, v ...interface{}) {
	std.Output(2, LDEBUG, fmt.Sprintf(f, v...))
}

func Info(v ...interface{}) {
	std.Output(2, LINFO, fmt.Sprint(v...))
}

func Infoln(v ...interface{}) {
	std.Output(2, LINFO, fmt.Sprintln(v...))
}

func Infof(f string, v ...interface{}) {
	std.Output(2, LINFO, fmt.Sprintf(f, v...))
}

func Error(v ...interface{}) {
	std.Output(2, LERROR, fmt.Sprint(v...))
}

func Errorln(v ...interface{}) {
	std.Output(2, LERROR, fmt.Sprintln(v...))
}

func Errorf(f string, v ...interface{}) {
	std.Output(2, LERROR, fmt.Sprintf(f, v...))
}

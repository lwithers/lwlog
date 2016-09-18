package lwlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	// DefaultTimeFormat is used to format timestamps. It is ISO8601
	// (extended form) with fractional second to µsecond precision.
	DefaultTimeFormat = "2006-01-02T15:04:05.000000Z07:00"
)

// Std is the standard logger implementation. It writes line-by-line output to
// a file descriptor.
type Std struct {
	// Debug allows debug messages to be switched on/off as desired
	Debug bool

	// settings
	timeFormat string
	out, err   io.Writer

	// cache of function addresses to string details, for fast lookup
	funcs     map[uintptr]*stdLoggerFuncDetails
	funcsLock sync.RWMutex
}

type stdLoggerFuncDetails struct {
	filename, function string
}

// NewStd returns a new standard logger. Timestamps are logged using
// DefaultTimeFormat. Debug messages are turned off by default, but will be
// logged to stdout if switched on. Info messages are logged to stdout. Error
// messages are logged to os.Stderr.
func NewStd() *Std {
	return &Std{
		timeFormat: DefaultTimeFormat,
		out:        os.Stdout,
		err:        os.Stderr,
		funcs:      make(map[uintptr]*stdLoggerFuncDetails),
	}
}

// NewLogger returns a new logging object, allowing the time format and output
// writers to be specified. The writers may be the same.
func NewLogger(timeFormat string, logOut, errOut io.Writer) *Std {
	return &Std{
		timeFormat: timeFormat,
		out:        logOut,
		err:        errOut,
		funcs:      make(map[uintptr]*stdLoggerFuncDetails),
	}
}

// Debugf logs a debug message. Debug messages may be switched on or off at
// runtime by setting l.Debug to true/false.
func (l *Std) Debugf(fmt string, args ...interface{}) {
	if !l.Debug {
		return
	}
	l.logLine(l.out, "debug", fmt, args...)
}

// Infof logs a message.
func (l *Std) Infof(fmt string, args ...interface{}) {
	l.logLine(l.out, "info ", fmt, args...)
}

// Errorf logs an error message.
func (l *Std) Errorf(fmt string, args ...interface{}) {
	l.logLine(l.err, "error", fmt, args...)
}

func (l *Std) logLine(w io.Writer, level, Fmt string, args ...interface{}) {
	// walk back over the stack (at most 4 entries) until we hit the first
	// non-logging function
	pc := make([]uintptr, 4)
	runtime.Callers(3, pc) // skip stdLogLine and Std.Debugf/Infof/etc.
	frames := runtime.CallersFrames(pc)
	frame, moreFrames := frames.Next()
	for moreFrames && IsLoggingFunction(frame.Func) {
		frame, moreFrames = frames.Next()
	}

	// grab the function name and filename, possibly using our cache
	filename, function := l.lookupFunc(frame.Func)

	// build the log message into an in-memory buffer, so that we don't
	// end up with interleaved output
	buf := bytes.NewBuffer(make([]byte, 0, 80))
	fmt.Fprintf(buf, "%s [%s] %s:(%s:%d): ",
		time.Now().Format(l.timeFormat),
		level, function, filename, frame.Line)

	fmt.Fprintf(buf, Fmt, args...)

	// test that the line ended with a newline, and add one if necessary
	n := len(buf.Bytes())
	if buf.Bytes()[n-1] != '\n' {
		_ = buf.WriteByte('\n')
	}

	// output in one chunk
	_, _ = w.Write(buf.Bytes())
}

func (l *Std) lookupFunc(f *runtime.Func) (filename, function string) {
	// we'll use the function's entrypoint address as the key in our cache
	pc := f.Entry()

	// test whether we've already cached the answer for this function
	l.funcsLock.RLock()
	details, known := l.funcs[pc]
	l.funcsLock.RUnlock()
	if known {
		return details.filename, details.function
	}

	// no cache entry — so extract the relevant details
	details = new(stdLoggerFuncDetails)

	// don't bother with the directory since the package name implies it
	file, _ := f.FileLine(pc)
	details.filename = filepath.Base(file)

	// only need the trailing path component of the package; that should
	// be enough to make it unambiguous in general without too much text
	funcname := f.Name()
	details.function = filepath.Join(filepath.Base(filepath.Dir(funcname)),
		filepath.Base(funcname))

	// update the cache (NB: it's possible that another goroutine could
	// also have passed into or through the above lock in the meantime,
	// but that's fine — it will just write the same result into the map).
	l.funcsLock.Lock()
	l.funcs[pc] = details
	l.funcsLock.Unlock()

	return details.filename, details.function
}

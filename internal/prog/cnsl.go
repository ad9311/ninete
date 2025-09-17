package prog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	reset  = "\x1b[0m"
	red    = "\x1b[31m"
	yellow = "\x1b[33m"
	blue   = "\x1b[34m"
	bold   = "\x1b[1m"
)

type Logger struct {
	mutex       sync.Mutex
	out         io.Writer
	outErr      io.Writer
	enableColor bool
	enableQuery bool
}

type LoggerOptions struct {
	Out         io.Writer
	OutErr      io.Writer
	EnableColor bool
	EnableQuery bool
}

func NewLogger(ops LoggerOptions) *Logger {
	if ops.Out == nil {
		ops.Out = os.Stdout
	}

	if ops.OutErr == nil {
		ops.OutErr = os.Stderr
	}

	return &Logger{
		out:         ops.Out,
		outErr:      ops.OutErr,
		enableColor: ops.EnableColor,
		enableQuery: ops.EnableQuery,
	}
}

func (l *Logger) Log(a any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := output(l.out, "%v", a); err != nil {
		panic(err)
	}
}

func (l *Logger) Logf(format string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := output(l.out, format, args...); err != nil {
		panic(err)
	}
}

func (l *Logger) Error(a any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	format := l.handleColor(red+"%v", "%v")

	if err := output(l.outErr, format, a); err != nil {
		panic(err)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	format = l.handleColor(red+format, format)

	if err := output(l.outErr, format, args...); err != nil {
		panic(err)
	}
}

func (l *Logger) Debug(a any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	format := l.handleColor(yellow+"%v", "%v")

	if err := output(l.out, format, a); err != nil {
		panic(err)
	}
}

func (l *Logger) Debugf(format string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	format = l.handleColor(yellow+format, format)

	if err := output(l.out, format, args...); err != nil {
		panic(err)
	}
}

func (l *Logger) Query(query string, dur time.Duration) {
	if !l.enableQuery {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	query = strings.TrimSpace(strings.ReplaceAll(query, "\n", " "))

	durStr := " [" + dur.String() + "]"
	format := bold + blue + query + ";" + red + durStr
	format = l.handleColor(format, query+";"+durStr)

	if err := output(l.out, format); err != nil {
		panic(err)
	}
}

func (l *Logger) handleColor(withColor, noColor string) string {
	if l.enableColor {
		return withColor + reset
	}

	return noColor
}

func output(w io.Writer, format string, args ...any) error {
	format = timestamp() + " " + format + "\n"
	_, err := fmt.Fprintf(w, format, args...)

	return err
}

func timestamp() string {
	return time.Now().UTC().Format("[02/01/06 15:04]")
}

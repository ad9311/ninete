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
	Out         io.Writer
	OutErr      io.Writer
	EnableColor bool
	EnableQuery bool

	mutex sync.Mutex
}

type LogOptions struct {
	Out         io.Writer
	OutErr      io.Writer
	EnableColor bool
	EnableQuery bool
}

func NewLogger(opt LogOptions) *Logger {
	if opt.Out == nil {
		opt.Out = os.Stdout
	}

	if opt.OutErr == nil {
		opt.OutErr = os.Stderr
	}

	return &Logger{
		Out:         opt.Out,
		OutErr:      opt.OutErr,
		EnableColor: opt.EnableColor,
		EnableQuery: opt.EnableQuery,
	}
}

func QuickLogger() *Logger {
	return &Logger{
		Out:         os.Stdout,
		OutErr:      os.Stderr,
		EnableColor: true,
		EnableQuery: true,
	}
}

func (l *Logger) Log(a any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := output(l.Out, "%v", a); err != nil {
		panic(err)
	}
}

func (l *Logger) Logf(format string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := output(l.Out, format, args...); err != nil {
		panic(err)
	}
}

func (l *Logger) Error(a any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	format := l.handleColor(red+"%v", "%v")

	if err := output(l.OutErr, format, a); err != nil {
		panic(err)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	colorFormat := l.handleColor(red+format, format)

	if err := output(l.OutErr, colorFormat, args...); err != nil {
		panic(err)
	}
}

func (l *Logger) Debug(a any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	format := l.handleColor(yellow+"%v", "%v")

	if err := output(l.Out, format, a); err != nil {
		panic(err)
	}
}

func (l *Logger) Debugf(format string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	colorFormat := l.handleColor(yellow+format, format)

	if err := output(l.Out, colorFormat, args...); err != nil {
		panic(err)
	}
}

func (l *Logger) Query(query string, dur time.Duration) {
	if !l.EnableQuery {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	query = strings.TrimSpace(strings.ReplaceAll(query, "\n", " "))

	durStr := " [" + dur.String() + "]"
	format := bold + blue + query + ";" + red + durStr
	format = l.handleColor(format, query+";"+durStr)

	if err := output(l.Out, format); err != nil {
		panic(err)
	}
}

func (l *Logger) handleColor(withColor, noColor string) string {
	if l.EnableColor {
		return withColor + reset
	}

	return noColor
}

func output(w io.Writer, format string, args ...any) error {
	finalFormat := timestamp() + " " + format + "\n"
	_, err := fmt.Fprintf(w, finalFormat, args...)

	return err
}

func timestamp() string {
	return time.Now().UTC().Format("[02/01/06 15:04]")
}

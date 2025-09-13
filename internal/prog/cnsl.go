package prog

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	yellow = "\x1b[33m"
	red    = "\x1b[31m"
	reset  = "\x1b[0m"
)

type semantic int

const (
	infoLevel semantic = iota
	errorLevel
	debugLevel
)

type Logger struct {
	mutex  sync.Mutex
	out    io.Writer
	outErr io.Writer
}

func NewLogger() *Logger {
	return &Logger{
		out:    os.Stdout,
		outErr: os.Stderr,
	}
}

func (l *Logger) Log(msg string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := writeLine(l.out, infoLevel, msg, args...); err != nil {
		panic(err)
	}
}

func (l *Logger) Error(msg string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := writeLine(l.outErr, errorLevel, msg, args...); err != nil {
		panic(err)
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := writeLine(l.out, debugLevel, msg, args...); err != nil {
		panic(err)
	}
}

func writeLine(w io.Writer, level semantic, msg string, args ...any) error {
	var body string

	switch level {
	case errorLevel:
		body = red + msg + reset
	case debugLevel:
		body = yellow + msg + reset
	default:
		body = msg
	}

	line := timestamp() + " " + body + "\n"
	_, err := fmt.Fprintf(w, line, args...)

	return err
}

func timestamp() string {
	return time.Now().UTC().Format("[02/01/06-15:04]")
}

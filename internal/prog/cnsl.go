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

type semantic int

const (
	infoLevel semantic = iota
	errorLevel
	debugLevel
	queryLevel
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

func (l *Logger) Query(query string, dur time.Duration) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	query = strings.TrimSpace(strings.ReplaceAll(query, "\n", " "))

	msg := blue + query + red + " [" + dur.String() + "]"

	if err := writeLine(l.out, queryLevel, msg); err != nil {
		panic(err)
	}
}

func writeLine(w io.Writer, level semantic, msg string, args ...any) error {
	var body string

	switch level {
	case errorLevel:
		body = red + msg
	case debugLevel:
		body = yellow + msg
	case queryLevel:
		body = bold + msg
	default:
		body = msg
	}

	line := timestamp() + " " + body + reset + "\n"
	_, err := fmt.Fprintf(w, line, args...)

	return err
}

func timestamp() string {
	return time.Now().UTC().Format("[02/01/06-15:04]")
}

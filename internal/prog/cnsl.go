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

// Logger provides a thread-safe logging mechanism with separate output streams
// for standard and error messages.
type Logger struct {
	mutex  sync.Mutex
	out    io.Writer
	outErr io.Writer
}

// NewLogger creates and returns a new Logger instance with standard output and error streams.
func NewLogger() *Logger {
	return &Logger{
		out:    os.Stdout,
		outErr: os.Stderr,
	}
}

// Log formats and writes a log message to the output stream in a thread-safe manner.
func (l *Logger) Log(msg string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := writeLine(l.out, infoLevel, msg, args...); err != nil {
		panic(err)
	}
}

// Error logs an error message to the standard error output in a thread-safe manner.
func (l *Logger) Error(msg string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := writeLine(l.outErr, errorLevel, msg, args...); err != nil {
		panic(err)
	}
}

// Debug logs a formatted debug message to the output stream in a thread-safe manner.
func (l *Logger) Debug(msg string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := writeLine(l.out, debugLevel, msg, args...); err != nil {
		panic(err)
	}
}

// writeLine writes a formatted log message to the provided io.Writer.
// Returns an error if writing to the writer fails.
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

// timestamp returns the current UTC time formatted as [DD/MM/YY-HH:MM].
func timestamp() string {
	return time.Now().UTC().Format("[02/01/06-15:04]")
}

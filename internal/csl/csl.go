// Package csl provides a simple, thread-safe logger
// with color-coded output for logs, errors, and debug messages.
package csl

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ad9311/ninete/internal/errs"
)

// Colors for the logger
const (
	yellow = "\x1b[33m"
	red    = "\x1b[31m"
	reset  = "\x1b[0m"
)

// Logger is a thread-safe logger with color-coded output for logs, errors, and debug messages.
type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	errOut  io.Writer
	QueryOn bool
}

// New creates and returns a new Logger instance with the specified output and error output writers.
// Both out and errOut must not be nil.
func New(out, errOut io.Writer) (*Logger, error) {
	var l *Logger

	if out == nil || errOut == nil {
		return l, fmt.Errorf("%w for writers", errs.ErrInterfaceNotSet)
	}

	l = &Logger{
		out:     out,
		errOut:  errOut,
		QueryOn: true,
	}

	return l, nil
}

// Log formats and writes a log message to the logger's output.
func (l *Logger) Log(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := l.write(l.out, "log", msg); err != nil {
		panic(err)
	}
}

// Error formats and writes an error message to the logger's output.
func (l *Logger) Error(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := l.write(l.errOut, "error", msg); err != nil {
		panic(err)
	}
}

// Debug formats and writes a debug message to the logger's output.
func (l *Logger) Debug(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := l.write(l.out, "debug", msg); err != nil {
		panic(err)
	}
}

// NewLog creates a new logger instance and logs a formatted message.
func NewLog(msg string, args ...any) {
	l, err := New(os.Stdout, os.Stderr)
	if err != nil {
		panic(err)
	}

	l.Log(msg, args...)
}

// NewError creates a new logger instance and logs an error formatted message.
func NewError(msg string, args ...any) {
	l, err := New(os.Stdout, os.Stderr)
	if err != nil {
		panic(err)
	}

	l.Error(msg, args...)
}

// NewDebug creates a new logger instance and logs a debug formatted message.
func NewDebug(msg string, args ...any) {
	l, err := New(os.Stdout, os.Stderr)
	if err != nil {
		panic(err)
	}

	l.Debug(msg, args...)
}

// ts returns the current UTC time formatted as a string in the pattern "[dd/mm/yy-hh:mm]".
func (l *Logger) ts() string {
	return time.Now().UTC().Format("[02/01/06-15:04]")
}

// write formats and writes a log message to the provided io.Writer.
// The output is prefixed with a timestamp and colored based on the log kind.
// Supported kinds are "log" (plain), "error" (red), and "debug" (yellow).
// Returns an error if writing to the writer fails.
func (l *Logger) write(w io.Writer, kind, msg string) error {
	var body string

	switch kind {
	case "log":
		body = msg
	case "error":
		body = red + msg + reset
	case "debug":
		body = yellow + msg + reset
	default:
		body = msg
	}

	line := l.ts() + " " + body + "\n"
	_, err := io.WriteString(w, line)

	return err
}

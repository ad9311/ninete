// Package csl provides a simple, thread-safe logger with color and timestamp support.
package csl

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

// Logger provides thread-safe logging capabilities with support for debug mode,
// colored output, and separate writers for standard and error outputs.
type Logger struct {
	mu       sync.Mutex // Mutex to ensure thread-safe logging
	out      io.Writer  // Writer for standard output
	errOut   io.Writer  // Writer for error output
	DebugOn  bool       // Enables debug logging if true
	UseColor bool       // Enables colored output if true
}

// New creates a new Logger instance.
// If out or errOut are nil, they default to os.Stdout and os.Stderr respectively.
// The debug flag enables debug logging.
func New(out, errOut io.Writer, debug bool) *Logger {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}

	return &Logger{
		out:      out,
		errOut:   errOut,
		DebugOn:  debug,
		UseColor: true,
	}
}

// ts returns the current UTC timestamp in [DD/MM/YY-HH:MM] format.
func (l *Logger) ts() string {
	return time.Now().UTC().Format("[02/01/06-15:04]")
}

// write writes the string s to the given writer w.
func (l *Logger) write(w io.Writer, s string) error {
	_, err := io.WriteString(w, s)

	return err
}

// Log writes a formatted log message with a timestamp to the standard output.
func (l *Logger) Log(msg string, args ...any) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	line := fmt.Sprintf("%s %s\n", l.ts(), fmt.Sprintf(msg, args...))

	return l.write(l.out, line)
}

// Debug writes a formatted debug message with a timestamp to the standard output if DebugOn is true.
// The message is colored yellow if UseColor is enabled.
func (l *Logger) Debug(msg string, args ...any) error {
	if !l.DebugOn {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	body := fmt.Sprintf(msg, args...)
	if l.UseColor {
		body = yellow + body + reset
	}
	line := fmt.Sprintf("%s %s\n", l.ts(), body)

	return l.write(l.out, line)
}

// Error writes a formatted error message with a timestamp to the error output.
// The message is colored red if UseColor is enabled.
func (l *Logger) Error(msg string, args ...any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	body := fmt.Sprintf(msg, args...)
	if l.UseColor {
		body = red + body + reset
	}
	line := fmt.Sprintf("%s %s\n", l.ts(), body)

	return l.write(l.errOut, line)
}

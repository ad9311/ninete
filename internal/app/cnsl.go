package app

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Colors for the logger
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

//nolint:gochecknoglobals
var (
	mutex sync.Mutex

	out    io.Writer = os.Stdout
	outErr io.Writer = os.Stderr
)

//nolint:gochecknoglobals

// Log formats and writes a log message to the output stream in a thread-safe manner.
func Log(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := writeLine(out, infoLevel, msg, args...); err != nil {
		panic(err)
	}
}

// LogError logs an error message to the standard error output in a thread-safe manner.
func LogError(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := writeLine(outErr, errorLevel, msg, args...); err != nil {
		panic(err)
	}
}

// Debug logs a formatted debug message to the output stream in a thread-safe manner.
func Debug(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := writeLine(out, debugLevel, msg, args...); err != nil {
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

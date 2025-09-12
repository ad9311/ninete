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

var (
	mutex sync.Mutex

	out    io.Writer = os.Stdout
	outErr io.Writer = os.Stderr
)

// Log formats and writes a log message to the output stream in a thread-safe manner.
func Log(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := writeLine(out, "log", msg); err != nil {
		panic(err)
	}
}

// LogError logs an error message to the standard error output in a thread-safe manner.
func LogError(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := writeLine(outErr, "error", msg); err != nil {
		panic(err)
	}
}

// Debug logs a formatted debug message to the output stream in a thread-safe manner.
func Debug(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := writeLine(out, "debug", msg); err != nil {
		panic(err)
	}
}

// writeLine writes a formatted log message to the provided io.Writer.
// Returns an error if writing to the writer fails.
func writeLine(w io.Writer, kind, msg string) error {
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

	line := timestamp() + " " + body + "\n"
	_, err := io.WriteString(w, line)

	return err
}

// timestamp returns the current UTC time formatted as [DD/MM/YY-HH:MM].
func timestamp() string {
	return time.Now().UTC().Format("[02/01/06-15:04]")
}

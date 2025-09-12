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

func Log(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := writeLine(out, "log", msg); err != nil {
		panic(err)
	}
}

func LogError(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := writeLine(outErr, "error", msg); err != nil {
		panic(err)
	}
}

func LogDebug(msg string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()

	msg = fmt.Sprintf(msg, args...)

	if err := writeLine(out, "debug", msg); err != nil {
		panic(err)
	}
}

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

func timestamp() string {
	return time.Now().UTC().Format("[02/01/06-15:04]")
}

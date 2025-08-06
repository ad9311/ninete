package task

import (
	"bytes"
	"log"
)

// captureLogOutput temporarily redirects the standard logger's output to a buffer,
// executes the provided function f, and returns the captured log output as a string.
// It restores the original logger output after execution.
func captureLogOutput(f func()) string {
	var buf bytes.Buffer

	oldWriter := log.Writer()
	defer log.SetOutput(oldWriter)

	log.SetOutput(&buf)

	f()

	return buf.String()
}

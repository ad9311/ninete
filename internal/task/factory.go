package task

import (
	"bytes"
	"io"
	"testing"

	"github.com/ad9311/go-api-base/internal/console"
)

// testBuffer is a helper struct that provides separate buffers for capturing
// standard output (stdOut) and standard error (stdErr) streams, typically used
// in testing scenarios to verify output.
type testBuffer struct {
	stdErr bytes.Buffer
	stdOut bytes.Buffer
}

// newTaskFactory is a helper function for tests that creates and returns a new *task instance.
// It calls setUp to initialize the task and fails the test immediately if an error occurs during setup.
func newTaskFactory(t *testing.T, out, errOut io.Writer) *task {
	task, err := setUp()

	logger := console.New(out, errOut, true)
	task.logger = logger

	if err != nil {
		t.Fatalf("failed to set up tasks: %v", err)
	}

	return task
}

// newTestBuffer creates and returns a pointer to a new testBuffer instance
// with initialized standard output and error buffers.
func newTestBuffer() *testBuffer {
	var bufOut bytes.Buffer
	var bufErr bytes.Buffer

	return &testBuffer{
		stdOut: bufOut,
		stdErr: bufErr,
	}
}

// clear resets the testBuffer by clearing its stdOut and stdErr buffers.
func (t *testBuffer) clear() {
	t.stdOut = *bytes.NewBuffer([]byte{})
	t.stdErr = *bytes.NewBuffer([]byte{})
}

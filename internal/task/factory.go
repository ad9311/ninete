package task

import "testing"

// newTaskFactory is a helper function for tests that creates and returns a new *task instance.
// It calls setUp to initialize the task and fails the test immediately if an error occurs during setup.
func newTaskFactory(t *testing.T) *task {
	task, err := setUp()
	if err != nil {
		t.Fatalf("failed to set up tasks: %v", err)
	}

	return task
}

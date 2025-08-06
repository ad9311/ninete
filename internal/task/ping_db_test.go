package task

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPingDatabaseTask(t *testing.T) {
	task := newTaskFactory(t)

	err := task.pingDatabaseTask()
	require.Nil(t, err)
}

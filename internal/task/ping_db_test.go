package task

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPingDatabaseTask(t *testing.T) {
	task := newTaskFactory(t, nil, nil)

	err := task.pingDatabaseTask()
	require.Nil(t, err)
}

package task

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDeleteExpiredTokensTask(t *testing.T) {
	ctx := context.Background()
	tbuff := newTestBuffer()
	task := newTaskFactory(t, &tbuff.stdOut, &tbuff.stdErr)

	user := task.serviceStore.FactoryUser(ctx, t, service.RegistrationParams{})

	for range 5 {
		task.serviceStore.FactoryExpiredToken(ctx, t, user.ID)
	}

	err := task.deleteExpiredTokensTask()
	require.Nil(t, err)

	want := "deleted 5 expired refresh tokens"
	require.Contains(t, tbuff.stdOut.String(), want)
}

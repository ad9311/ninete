package task

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDeleteExpiredTokensTask(t *testing.T) {
	ctx := context.Background()
	task := newTaskFactory(t)

	user := task.serviceStore.FactoryUser(ctx, t, service.RegistrationParams{})

	for range 5 {
		task.serviceStore.FactoryExpiredToken(ctx, t, user.ID)
	}

	got := captureLogOutput(func() {
		err := task.deleteExpiredTokensTask()
		require.Nil(t, err)
	})

	want := "deleted 5 expired refresh tokens"
	require.Contains(t, got, want)
}

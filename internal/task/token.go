package task

import (
	"context"
)

// deleteExpiredTokensTask deletes all expired refresh tokens and logs the result.
func (t *task) deleteExpiredTokensTask() error {
	ctx := context.Background()
	count, err := t.serviceStore.DeleteExpieredRefreshTokens(ctx)
	if err != nil {
		return err
	}

	t.logger.Log("deleted %d expired refresh tokens", count)

	return nil
}

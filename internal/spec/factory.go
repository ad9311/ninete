package spec

import (
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/stretchr/testify/require"
)

func (s *Spec) CreateUser(t *testing.T, params repo.InsertUserParams) logic.User {
	t.Helper()

	user, err := s.Store.CreateUser(t.Context(), params)
	require.NoError(t, err)

	return user
}

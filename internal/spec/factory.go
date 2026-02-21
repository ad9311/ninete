package spec

import (
	"database/sql"
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

func (s *Spec) CreateCategory(t *testing.T, name string) repo.Category {
	t.Helper()

	category, err := s.Store.CreateCategory(t.Context(), name)
	require.NoError(t, err)

	return category
}

func (s *Spec) CreateInvitationCode(t *testing.T, rawCode string) repo.InvitationCode {
	t.Helper()

	invitationCode, err := s.Store.CreateInvitationCode(t.Context(), rawCode)
	require.NoError(t, err)

	return invitationCode
}

func (s *Spec) CreateAuthUser(
	t *testing.T,
	username, email, rawPassword string,
) logic.User {
	t.Helper()

	passwordHash, err := logic.HashPassword(rawPassword)
	require.NoError(t, err)

	return s.CreateUser(t, repo.InsertUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	})
}

func (s *Spec) CreateTag(t *testing.T, userID int, name string) repo.Tag {
	t.Helper()

	tag, err := s.Store.CreateTag(t.Context(), userID, logic.TagParams{Name: name})
	require.NoError(t, err)

	return tag
}

func (s *Spec) CreateExpense(t *testing.T, userID int, params logic.ExpenseParams) repo.Expense {
	t.Helper()

	expense, err := s.Store.CreateExpense(t.Context(), userID, params)
	require.NoError(t, err)

	return expense
}

func (s *Spec) SetRecurrentExpenseLastCopy(
	t *testing.T,
	re repo.RecurrentExpense,
	lastCopy int64,
) repo.RecurrentExpense {
	t.Helper()

	updated, err := s.Queries.UpdateRecurrentExpense(t.Context(), repo.UpdateRecurrentExpenseParams{
		ID:                re.ID,
		UserID:            re.UserID,
		CategoryID:        re.CategoryID,
		Description:       re.Description,
		Amount:            re.Amount,
		Period:            re.Period,
		LastCopyCreatedAt: sql.NullInt64{Int64: lastCopy, Valid: true},
	})
	require.NoError(t, err)

	return updated
}

func (s *Spec) CreateRecurrentExpense(
	t *testing.T,
	userID int,
	params logic.RecurrentExpenseParams,
) repo.RecurrentExpense {
	t.Helper()

	recurrentExpense, err := s.Store.CreateRecurrentExpense(t.Context(), userID, params)
	require.NoError(t, err)

	return recurrentExpense
}

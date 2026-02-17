package logic_test

import (
	"database/sql"
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_user",
			fn: func(t *testing.T) {
				params := repo.InsertUserParams{
					Username:     "createuser",
					Email:        "createuser@example.com",
					PasswordHash: []byte("createuser_hash"),
				}

				user, err := s.Store.CreateUser(ctx, params)
				require.NoError(t, err)
				require.Positive(t, user.ID)
				require.Equal(t, params.Username, user.Username)
				require.Equal(t, params.Email, user.Email)
				require.NotZero(t, user.CreatedAt)
				require.NotZero(t, user.UpdatedAt)
			},
		},
		{
			name: "should_fail_with_duplicate_email",
			fn: func(t *testing.T) {
				baseEmail := "duplicate_email_@example.com"
				first := repo.InsertUserParams{
					Username:     "duplicateemail_first",
					Email:        baseEmail,
					PasswordHash: []byte("duplicateemail_hash_1"),
				}
				second := repo.InsertUserParams{
					Username:     "duplicateemail_second_",
					Email:        baseEmail,
					PasswordHash: []byte("duplicateemail_hash_2"),
				}

				_, err := s.Store.CreateUser(ctx, first)
				require.NoError(t, err)

				_, err = s.Store.CreateUser(ctx, second)
				require.Error(t, err)
			},
		},
		{
			name: "should_fail_with_duplicate_username",
			fn: func(t *testing.T) {
				baseUsername := "duplicate_username"
				first := repo.InsertUserParams{
					Username:     baseUsername,
					Email:        "duplicateusername_first@example.com",
					PasswordHash: []byte("duplicateusername_hash_1"),
				}
				second := repo.InsertUserParams{
					Username:     baseUsername,
					Email:        "duplicateusername_second@example.com",
					PasswordHash: []byte("duplicateusername_hash_2"),
				}

				_, err := s.Store.CreateUser(t.Context(), first)
				require.NoError(t, err)

				_, err = s.Store.CreateUser(t.Context(), second)
				require.Error(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindUser(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_user_by_id",
			fn: func(t *testing.T) {
				createdUser := s.CreateUser(t, repo.InsertUserParams{
					Username:     "finduser",
					Email:        "finduser@example.com",
					PasswordHash: []byte("finduser_hash"),
				})

				foundUser, err := s.Store.FindUser(ctx, createdUser.ID)
				require.NoError(t, err)
				require.Equal(t, createdUser.ID, foundUser.ID)
				require.Equal(t, createdUser.Username, foundUser.Username)
				require.Equal(t, createdUser.Email, foundUser.Email)
			},
		},
		{
			name: "should_fail_when_user_not_found",
			fn: func(t *testing.T) {
				_, err := s.Store.FindUser(ctx, -1)
				require.Error(t, err)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindUserForAuth(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_user_by_email_for_auth",
			fn: func(t *testing.T) {
				passwordHash := "finduserforauth_hash"
				params := repo.InsertUserParams{
					Username:     "finduserforauth",
					Email:        "finduserforauth@example.com",
					PasswordHash: []byte(passwordHash),
				}

				createdUser := s.CreateUser(t, params)

				foundUser, err := s.Store.FindUserForAuth(ctx, params.Email)
				require.NoError(t, err)
				require.Equal(t, createdUser.ID, foundUser.ID)
				require.Equal(t, params.Username, foundUser.Username)
				require.Equal(t, params.Email, foundUser.Email)
				require.Equal(t, []byte(passwordHash), foundUser.PasswordHash)
			},
		},
		{
			name: "should_fail_when_email_not_found",
			fn: func(t *testing.T) {
				_, err := s.Store.FindUserForAuth(
					t.Context(),
					"missing_auth_email@example.com",
				)
				require.Error(t, err)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

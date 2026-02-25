package logic_test

import (
	"database/sql"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateList(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "list_user_1",
		Email:        "list_user_1@example.com",
		PasswordHash: []byte("list_user_hash_1"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "list_user_2",
		Email:        "list_user_2@example.com",
		PasswordHash: []byte("list_user_hash_2"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_list",
			fn: func(t *testing.T) {
				list, err := s.Store.CreateList(ctx, user.ID, logic.ListParams{Name: "My List"})
				require.NoError(t, err)
				require.Positive(t, list.ID)
				require.Equal(t, user.ID, list.UserID)
				require.Equal(t, "My List", list.Name)
			},
		},
		{
			name: "should_fail_validation_when_name_is_empty",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateList(ctx, user.ID, logic.ListParams{Name: ""})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_validation_when_name_exceeds_max_length",
			fn: func(t *testing.T) {
				longName := "this list name is way too long and exceeds" +
					" the one hundred character maximum allowed for a list name!"
				_, err := s.Store.CreateList(ctx, user.ID, logic.ListParams{Name: longName})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_with_duplicate_name_for_same_user",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateList(ctx, user.ID, logic.ListParams{Name: "Duplicate List"})
				require.NoError(t, err)

				_, err = s.Store.CreateList(ctx, user.ID, logic.ListParams{Name: "Duplicate List"})
				require.Error(t, err)
			},
		},
		{
			name: "should_allow_same_name_for_different_users",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateList(ctx, user.ID, logic.ListParams{Name: "Shared Name"})
				require.NoError(t, err)

				otherList, err := s.Store.CreateList(ctx, otherUser.ID, logic.ListParams{Name: "Shared Name"})
				require.NoError(t, err)
				require.Equal(t, otherUser.ID, otherList.UserID)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateList(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "list_user_3",
		Email:        "list_user_3@example.com",
		PasswordHash: []byte("list_user_hash_3"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "list_user_4",
		Email:        "list_user_4@example.com",
		PasswordHash: []byte("list_user_hash_4"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_list_name",
			fn: func(t *testing.T) {
				list := s.CreateList(t, user.ID, "Original Name 1")

				updated, err := s.Store.UpdateList(ctx, list.ID, user.ID, logic.ListParams{Name: "Updated Name 1"})
				require.NoError(t, err)
				require.Equal(t, list.ID, updated.ID)
				require.Equal(t, "Updated Name 1", updated.Name)
			},
		},
		{
			name: "should_fail_when_updating_another_users_list",
			fn: func(t *testing.T) {
				list := s.CreateList(t, user.ID, "Original Name 2")

				_, err := s.Store.UpdateList(ctx, list.ID, otherUser.ID, logic.ListParams{Name: "Updated Name 2"})
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteList(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "list_user_5",
		Email:        "list_user_5@example.com",
		PasswordHash: []byte("list_user_hash_5"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "list_user_6",
		Email:        "list_user_6@example.com",
		PasswordHash: []byte("list_user_hash_6"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_list_for_owner",
			fn: func(t *testing.T) {
				list := s.CreateList(t, user.ID, "List To Delete 1")

				deletedID, err := s.Store.DeleteList(ctx, list.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, list.ID, deletedID)
			},
		},
		{
			name: "should_fail_when_deleting_another_users_list",
			fn: func(t *testing.T) {
				list := s.CreateList(t, user.ID, "List To Delete 2")

				_, err := s.Store.DeleteList(ctx, list.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindList(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "list_user_7",
		Email:        "list_user_7@example.com",
		PasswordHash: []byte("list_user_hash_7"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "list_user_8",
		Email:        "list_user_8@example.com",
		PasswordHash: []byte("list_user_hash_8"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_list_by_id_and_owner",
			fn: func(t *testing.T) {
				list := s.CreateList(t, user.ID, "Find Me")

				found, err := s.Store.FindList(ctx, list.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, list.ID, found.ID)
				require.Equal(t, list.Name, found.Name)
			},
		},
		{
			name: "should_not_find_list_for_wrong_user",
			fn: func(t *testing.T) {
				list := s.CreateList(t, user.ID, "Private List")

				_, err := s.Store.FindList(ctx, list.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

package logic_test

import (
	"database/sql"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateTask(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "task_user_1",
		Email:        "task_user_1@example.com",
		PasswordHash: []byte("task_user_hash_1"),
	})
	list := s.CreateList(t, user.ID, "Task Test List 1")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_task",
			fn: func(t *testing.T) {
				task, err := s.Store.CreateTask(ctx, list.ID, user.ID, logic.TaskParams{
					Description: "Buy groceries",
					Priority:    1,
				})
				require.NoError(t, err)
				require.Positive(t, task.ID)
				require.Equal(t, list.ID, task.ListID)
				require.Equal(t, user.ID, task.UserID)
				require.Equal(t, "Buy groceries", task.Description)
				require.Equal(t, 1, task.Priority)
				require.False(t, task.Done)
			},
		},
		{
			name: "should_create_task_with_tags",
			fn: func(t *testing.T) {
				task, err := s.Store.CreateTask(ctx, list.ID, user.ID, logic.TaskParams{
					Description: "Tagged task",
					Priority:    2,
					Tags:        []string{"work", "urgent"},
				})
				require.NoError(t, err)
				require.Positive(t, task.ID)

				tags, err := s.Store.FindTaskTags(ctx, task.ID, user.ID)
				require.NoError(t, err)
				require.Len(t, tags, 2)
			},
		},
		{
			name: "should_fail_validation_when_description_is_empty",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateTask(ctx, list.ID, user.ID, logic.TaskParams{
					Description: "",
					Priority:    1,
				})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_validation_when_priority_is_out_of_range",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateTask(ctx, list.ID, user.ID, logic.TaskParams{
					Description: "Bad priority",
					Priority:    4,
				})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_validation_when_priority_is_zero",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateTask(ctx, list.ID, user.ID, logic.TaskParams{
					Description: "Zero priority",
					Priority:    0,
				})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateTask(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "task_user_2",
		Email:        "task_user_2@example.com",
		PasswordHash: []byte("task_user_hash_2"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "task_user_3",
		Email:        "task_user_3@example.com",
		PasswordHash: []byte("task_user_hash_3"),
	})
	list := s.CreateList(t, user.ID, "Task Test List 2")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_task_fields",
			fn: func(t *testing.T) {
				task := s.CreateTask(t, list.ID, user.ID, logic.TaskParams{
					Description: "Original description",
					Priority:    1,
				})

				updated, err := s.Store.UpdateTask(ctx, task.ID, user.ID, logic.TaskParams{
					Description: "Updated description",
					Priority:    3,
					Done:        false,
				})
				require.NoError(t, err)
				require.Equal(t, task.ID, updated.ID)
				require.Equal(t, "Updated description", updated.Description)
				require.Equal(t, 3, updated.Priority)
			},
		},
		{
			name: "should_toggle_done_status",
			fn: func(t *testing.T) {
				task := s.CreateTask(t, list.ID, user.ID, logic.TaskParams{
					Description: "Toggle me",
					Priority:    2,
				})
				require.False(t, task.Done)

				updated, err := s.Store.UpdateTask(ctx, task.ID, user.ID, logic.TaskParams{
					Description: task.Description,
					Priority:    task.Priority,
					Done:        true,
				})
				require.NoError(t, err)
				require.True(t, updated.Done)
			},
		},
		{
			name: "should_replace_tags_on_update",
			fn: func(t *testing.T) {
				task := s.CreateTask(t, list.ID, user.ID, logic.TaskParams{
					Description: "Tag replacement task",
					Priority:    1,
					Tags:        []string{"old_tag"},
				})

				_, err := s.Store.UpdateTask(ctx, task.ID, user.ID, logic.TaskParams{
					Description: task.Description,
					Priority:    task.Priority,
					Tags:        []string{"new_tag_a", "new_tag_b"},
				})
				require.NoError(t, err)

				tags, err := s.Store.FindTaskTags(ctx, task.ID, user.ID)
				require.NoError(t, err)
				require.Len(t, tags, 2)
				require.Equal(t, "new_tag_a", tags[0].Name)
				require.Equal(t, "new_tag_b", tags[1].Name)
			},
		},
		{
			name: "should_fail_when_updating_another_users_task",
			fn: func(t *testing.T) {
				task := s.CreateTask(t, list.ID, user.ID, logic.TaskParams{
					Description: "Protected task",
					Priority:    1,
				})

				_, err := s.Store.UpdateTask(ctx, task.ID, otherUser.ID, logic.TaskParams{
					Description: "Hacked",
					Priority:    1,
				})
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteTask(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "task_user_4",
		Email:        "task_user_4@example.com",
		PasswordHash: []byte("task_user_hash_4"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "task_user_5",
		Email:        "task_user_5@example.com",
		PasswordHash: []byte("task_user_hash_5"),
	})
	list := s.CreateList(t, user.ID, "Task Test List 3")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_task_for_owner",
			fn: func(t *testing.T) {
				task := s.CreateTask(t, list.ID, user.ID, logic.TaskParams{
					Description: "Delete me",
					Priority:    1,
				})

				deletedID, err := s.Store.DeleteTask(ctx, task.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, task.ID, deletedID)
			},
		},
		{
			name: "should_fail_when_deleting_another_users_task",
			fn: func(t *testing.T) {
				task := s.CreateTask(t, list.ID, user.ID, logic.TaskParams{
					Description: "Protected delete",
					Priority:    1,
				})

				_, err := s.Store.DeleteTask(ctx, task.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindTask(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "task_user_6",
		Email:        "task_user_6@example.com",
		PasswordHash: []byte("task_user_hash_6"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "task_user_7",
		Email:        "task_user_7@example.com",
		PasswordHash: []byte("task_user_hash_7"),
	})
	list := s.CreateList(t, user.ID, "Task Test List 4")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_task_by_id_and_owner",
			fn: func(t *testing.T) {
				task := s.CreateTask(t, list.ID, user.ID, logic.TaskParams{
					Description: "Find me",
					Priority:    2,
				})

				found, err := s.Store.FindTask(ctx, task.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, task.ID, found.ID)
				require.Equal(t, task.Description, found.Description)
			},
		},
		{
			name: "should_not_find_task_for_wrong_user",
			fn: func(t *testing.T) {
				task := s.CreateTask(t, list.ID, user.ID, logic.TaskParams{
					Description: "Private task",
					Priority:    1,
				})

				_, err := s.Store.FindTask(ctx, task.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

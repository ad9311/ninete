package logic_test

import (
	"database/sql"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateMoodEntry(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "mood_user_1",
		Email:        "mood_user_1@example.com",
		PasswordHash: []byte("mood_user_hash_1"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_mood_entry",
			fn: func(t *testing.T) {
				params := newMoodEntryParams("Happy", "feeling great", 1735689600, nil)
				entry, err := s.Store.CreateMoodEntry(ctx, user.ID, params)
				require.NoError(t, err)
				require.Positive(t, entry.ID)
				require.Equal(t, user.ID, entry.UserID)
				require.Equal(t, "Happy", entry.Mood)
				require.Equal(t, "feeling great", entry.Notes)
			},
		},
		{
			name: "should_create_mood_entry_with_normalized_tags",
			fn: func(t *testing.T) {
				entry, err := s.Store.CreateMoodEntry(
					ctx,
					user.ID,
					newMoodEntryParams("Calm", "with tags", 1735776000, []string{"TAG_M_1", " tag_m_1 ", "tag_m_2"}),
				)
				require.NoError(t, err)

				tags, err := s.Store.FindMoodEntryTags(ctx, entry.ID, user.ID)
				require.NoError(t, err)
				require.Len(t, tags, 2)
				require.Equal(t, "tag_m_1", tags[0].Name)
				require.Equal(t, "tag_m_2", tags[1].Name)
			},
		},
		{
			name: "should_fail_for_invalid_mood",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateMoodEntry(ctx, user.ID, newMoodEntryParams("NotAMood", "", 1735862400, nil))
				require.ErrorIs(t, err, logic.ErrInvalidMood)
			},
		},
		{
			name: "should_fail_validation_for_missing_date",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateMoodEntry(ctx, user.ID, newMoodEntryParams("Happy", "", 0, nil))
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindMoodEntry(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "mood_user_2",
		Email:        "mood_user_2@example.com",
		PasswordHash: []byte("mood_user_hash_2"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "mood_user_3",
		Email:        "mood_user_3@example.com",
		PasswordHash: []byte("mood_user_hash_3"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_mood_entry",
			fn: func(t *testing.T) {
				created := s.CreateMoodEntry(t, user.ID, newMoodEntryParams("Excited", "test", 1735689600, nil))
				found, err := s.Store.FindMoodEntry(ctx, created.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, created.ID, found.ID)
				require.Equal(t, "Excited", found.Mood)
			},
		},
		{
			name: "should_not_find_other_users_entry",
			fn: func(t *testing.T) {
				created := s.CreateMoodEntry(t, user.ID, newMoodEntryParams("Sad", "private", 1735776000, nil))
				_, err := s.Store.FindMoodEntry(ctx, created.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateMoodEntry(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "mood_user_4",
		Email:        "mood_user_4@example.com",
		PasswordHash: []byte("mood_user_hash_4"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "mood_user_5",
		Email:        "mood_user_5@example.com",
		PasswordHash: []byte("mood_user_hash_5"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_mood_entry",
			fn: func(t *testing.T) {
				entry := s.CreateMoodEntry(t, user.ID, newMoodEntryParams("Tired", "before", 1735689600, nil))
				updateParams := newMoodEntryParams("Energized", "after", 1735776000, nil)
				updated, err := s.Store.UpdateMoodEntry(ctx, entry.ID, user.ID, updateParams)
				require.NoError(t, err)
				require.Equal(t, "Energized", updated.Mood)
				require.Equal(t, "after", updated.Notes)
			},
		},
		{
			name: "should_replace_tags_on_update",
			fn: func(t *testing.T) {
				entry := s.CreateMoodEntry(t, user.ID, newMoodEntryParams("Happy", "", 1735862400, []string{"old_tag"}))
				updateParams := newMoodEntryParams("Happy", "", 1735862400, []string{"new_tag"})
				_, err := s.Store.UpdateMoodEntry(ctx, entry.ID, user.ID, updateParams)
				require.NoError(t, err)

				tags, err := s.Store.FindMoodEntryTags(ctx, entry.ID, user.ID)
				require.NoError(t, err)
				require.Len(t, tags, 1)
				require.Equal(t, "new_tag", tags[0].Name)
			},
		},
		{
			name: "should_not_update_other_users_entry",
			fn: func(t *testing.T) {
				entry := s.CreateMoodEntry(t, user.ID, newMoodEntryParams("Calm", "", 1735948800, nil))
				_, err := s.Store.UpdateMoodEntry(ctx, entry.ID, otherUser.ID, newMoodEntryParams("Angry", "", 1735948800, nil))
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "should_fail_for_invalid_mood",
			fn: func(t *testing.T) {
				entry := s.CreateMoodEntry(t, user.ID, newMoodEntryParams("Calm", "", 1736035200, nil))
				updateParams := newMoodEntryParams("InvalidMood", "", 1736035200, nil)
				_, err := s.Store.UpdateMoodEntry(ctx, entry.ID, user.ID, updateParams)
				require.ErrorIs(t, err, logic.ErrInvalidMood)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteMoodEntry(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "mood_user_6",
		Email:        "mood_user_6@example.com",
		PasswordHash: []byte("mood_user_hash_6"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "mood_user_7",
		Email:        "mood_user_7@example.com",
		PasswordHash: []byte("mood_user_hash_7"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_mood_entry",
			fn: func(t *testing.T) {
				entry := s.CreateMoodEntry(t, user.ID, newMoodEntryParams("Bored", "", 1736121600, nil))
				err := s.Store.DeleteMoodEntry(ctx, entry.ID, user.ID)
				require.NoError(t, err)

				_, err = s.Store.FindMoodEntry(ctx, entry.ID, user.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "should_not_delete_other_users_entry",
			fn: func(t *testing.T) {
				entry := s.CreateMoodEntry(t, user.ID, newMoodEntryParams("Joyful", "", 1736208000, nil))
				err := s.Store.DeleteMoodEntry(ctx, entry.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func newMoodEntryParams(mood, notes string, loggedAt int64, tags []string) logic.MoodEntryParams {
	return logic.MoodEntryParams{
		Mood:     mood,
		Notes:    notes,
		LoggedAt: loggedAt,
		Tags:     tags,
	}
}

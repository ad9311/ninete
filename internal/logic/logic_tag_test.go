package logic_test

import (
	"database/sql"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateTag(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "tag_user_1",
		Email:        "tag_user_1@example.com",
		PasswordHash: []byte("tag_user_hash_1"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "tag_user_2",
		Email:        "tag_user_2@example.com",
		PasswordHash: []byte("tag_user_hash_2"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_tag_with_normalized_name",
			fn: func(t *testing.T) {
				tag, err := s.Store.CreateTag(ctx, user.ID, logic.TagParams{Name: " TAG_NAME_1 "})
				require.NoError(t, err)
				require.Positive(t, tag.ID)
				require.Equal(t, user.ID, tag.UserID)
				require.Equal(t, "tag_name_1", tag.Name)
			},
		},
		{
			name: "should_fail_validation_when_tag_name_is_empty",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateTag(ctx, user.ID, logic.TagParams{Name: "   "})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_validation_when_tag_name_is_too_long",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateTag(ctx, user.ID, logic.TagParams{Name: "this_tag_name_is_way_too_long"})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_with_duplicate_tag_for_same_user",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateTag(ctx, user.ID, logic.TagParams{Name: "tag_name_2"})
				require.NoError(t, err)

				_, err = s.Store.CreateTag(ctx, user.ID, logic.TagParams{Name: " TAG_NAME_2 "})
				require.Error(t, err)
			},
		},
		{
			name: "should_allow_same_tag_name_for_different_users",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateTag(ctx, user.ID, logic.TagParams{Name: "tag_name_3"})
				require.NoError(t, err)

				otherTag, err := s.Store.CreateTag(ctx, otherUser.ID, logic.TagParams{Name: "tag_name_3"})
				require.NoError(t, err)
				require.Equal(t, otherUser.ID, otherTag.UserID)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindTags(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "tag_user_3",
		Email:        "tag_user_3@example.com",
		PasswordHash: []byte("tag_user_hash_3"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "tag_user_4",
		Email:        "tag_user_4@example.com",
		PasswordHash: []byte("tag_user_hash_4"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_tags_for_filtered_user",
			fn: func(t *testing.T) {
				tagOne := s.CreateTag(t, user.ID, "tag_name_4")
				tagTwo := s.CreateTag(t, user.ID, "tag_name_5")
				s.CreateTag(t, otherUser.ID, "tag_name_6")

				tags, err := s.Store.FindTags(ctx, repo.QueryOptions{
					Filters: repo.Filters{
						FilterFields: []repo.FilterField{
							{Name: "user_id", Value: user.ID, Operator: "="},
						},
					},
					Sorting: repo.Sorting{
						Field: "name",
						Order: "ASC",
					},
				})
				require.NoError(t, err)
				require.Len(t, tags, 2)
				require.Equal(t, tagOne.ID, tags[0].ID)
				require.Equal(t, tagTwo.ID, tags[1].ID)
			},
		},
		{
			name: "should_fail_with_invalid_sort_field",
			fn: func(t *testing.T) {
				_, err := s.Store.FindTags(ctx, repo.QueryOptions{
					Sorting: repo.Sorting{
						Field: "invalid_field",
						Order: "ASC",
					},
				})
				require.ErrorIs(t, err, repo.ErrInvalidField)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteTag(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "tag_user_5",
		Email:        "tag_user_5@example.com",
		PasswordHash: []byte("tag_user_hash_5"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "tag_user_6",
		Email:        "tag_user_6@example.com",
		PasswordHash: []byte("tag_user_hash_6"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_tag_for_owner",
			fn: func(t *testing.T) {
				tag := s.CreateTag(t, user.ID, "tag_name_7")

				deletedID, err := s.Store.DeleteTag(ctx, tag.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, tag.ID, deletedID)
			},
		},
		{
			name: "should_fail_when_deleting_tag_of_another_user",
			fn: func(t *testing.T) {
				tag := s.CreateTag(t, user.ID, "tag_name_8")

				_, err := s.Store.DeleteTag(ctx, tag.ID, otherUser.ID)
				require.Error(t, err)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestParseTagNames(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_parse_normalize_and_dedupe_tag_names",
			fn: func(t *testing.T) {
				tagNames := logic.ParseTagNames("TAG_A; tag_a ; ; tag_b; tag_c ;tag_b")
				require.Equal(t, []string{"tag_a", "tag_b", "tag_c"}, tagNames)
			},
		},
		{
			name: "should_return_empty_slice_when_all_tags_are_blank",
			fn: func(t *testing.T) {
				tagNames := logic.ParseTagNames(" ; ; ")
				require.Empty(t, tagNames)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestJoinTagNames(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_join_tag_names_with_semicolon_and_space",
			fn: func(t *testing.T) {
				raw := logic.JoinTagNames([]string{"tag_a", "tag_b", "tag_c"})
				require.Equal(t, "tag_a; tag_b; tag_c", raw)
			},
		},
		{
			name: "should_return_empty_string_for_empty_input",
			fn: func(t *testing.T) {
				raw := logic.JoinTagNames([]string{})
				require.Equal(t, "", raw)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

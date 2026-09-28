package logic_test

import (
	"database/sql"
	"testing"
	"time"

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

func TestRetag(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "retag_user_1",
		Email:        "retag_user_1@example.com",
		PasswordHash: []byte("retag_user_hash_1"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "retag_user_2",
		Email:        "retag_user_2@example.com",
		PasswordHash: []byte("retag_user_hash_2"),
	})
	category := s.CreateCategory(t, "retag_category")

	expenseTags := func(t *testing.T, userID, expenseID int) []string {
		t.Helper()

		tags, err := s.Store.FindExpenseTags(ctx, expenseID, userID)
		require.NoError(t, err)

		return logic.ExtractTagNames(tags)
	}

	userTagNames := func(t *testing.T, userID int) []string {
		t.Helper()

		tags, err := s.Store.FindTags(ctx, repo.QueryOptions{
			Filters: repo.Filters{
				FilterFields: []repo.FilterField{
					{Name: "user_id", Value: userID, Operator: "="},
				},
			},
		})
		require.NoError(t, err)

		return logic.ExtractTagNames(tags)
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_rename_by_retagging_to_a_new_tag_and_keep_the_old_one",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(t, user.ID, newExpenseParams(
					category.ID, "retag rename", 100, 1735689600, []string{"rt_old", "rt_keep"}))

				result, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{" RT_OLD "},
					To:   " RT_NEW ",
				})
				require.NoError(t, err)
				require.Equal(t, "rt_new", result.Tag.Name)
				require.Equal(t, 1, result.Retagged)

				require.ElementsMatch(t, []string{"rt_new", "rt_keep"}, expenseTags(t, user.ID, expense.ID))
				require.Contains(t, userTagNames(t, user.ID), "rt_old",
					"the source tag was deleted; a retag must leave it in place")
			},
		},
		{
			name: "should_merge_several_tags_into_an_existing_one",
			fn: func(t *testing.T) {
				s.CreateTag(t, user.ID, "rt_merge_to")
				both := s.CreateExpense(t, user.ID, newExpenseParams(
					category.ID, "retag both", 100, 1735689600, []string{"rt_merge_a", "rt_merge_b"}))
				already := s.CreateExpense(t, user.ID, newExpenseParams(
					category.ID, "retag already", 100, 1735689600, []string{"rt_merge_a", "rt_merge_to"}))
				untouched := s.CreateExpense(t, user.ID, newExpenseParams(
					category.ID, "retag untouched", 100, 1735689600, []string{"rt_merge_other"}))

				result, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{"rt_merge_a", "rt_merge_b"},
					To:   "rt_merge_to",
				})
				require.NoError(t, err)
				require.Equal(t, 2, result.Retagged, "a record carrying two sources counts once")

				require.Equal(t, []string{"rt_merge_to"}, expenseTags(t, user.ID, both.ID))
				require.Equal(t, []string{"rt_merge_to"}, expenseTags(t, user.ID, already.ID))
				require.Equal(t, []string{"rt_merge_other"}, expenseTags(t, user.ID, untouched.ID))
			},
		},
		{
			name: "should_retag_recurrent_expenses",
			fn: func(t *testing.T) {
				params := newRecurrentExpenseParams(category.ID, "retag recurrent", 100, 1)
				params.Tags = []string{"rt_recurrent_from"}
				recurrent := s.CreateRecurrentExpense(t, user.ID, params)

				_, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{"rt_recurrent_from"},
					To:   "rt_recurrent_to",
				})
				require.NoError(t, err)

				tags, err := s.Store.FindRecurrentExpenseTags(ctx, recurrent.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, []string{"rt_recurrent_to"}, logic.ExtractTagNames(tags))
			},
		},
		{
			name: "should_move_the_report_grouping_onto_the_target",
			fn: func(t *testing.T) {
				from := s.CreateTag(t, user.ID, "rt_report_from")
				kept := s.CreateTag(t, user.ID, "rt_report_kept")
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{from.ID, kept.ID},
				}))

				result, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{"rt_report_from"},
					To:   "rt_report_to",
				})
				require.NoError(t, err)

				setting, err := s.Store.FindReportSetting(ctx, user.ID)
				require.NoError(t, err)
				require.ElementsMatch(t, []int{kept.ID, result.Tag.ID}, setting.TagIDs)
			},
		},
		{
			// A genuine reproduction: copying the taggings instead of updating
			// them in place gives the moved tag a newer tagging, and the report
			// then files the expense under rt_order_b instead.
			name: "should_keep_the_report_section_a_moved_tag_had",
			fn: func(t *testing.T) {
				first := s.CreateTag(t, user.ID, "rt_order_a")
				second := s.CreateTag(t, user.ID, "rt_order_b")
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{first.ID, second.ID},
				}))

				// March 2024, a month no other case bills to.
				month := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)
				s.CreateExpense(t, user.ID, newExpenseParams(
					category.ID, "retag order", 100, month.Unix(), []string{"rt_order_a", "rt_order_b"}))

				_, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{"rt_order_a"},
					To:   "rt_order_c",
				})
				require.NoError(t, err)

				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, month)
				require.NoError(t, err)
				require.Len(t, report.Sections, 1)
				require.Equal(t, "rt_order_c", report.Sections[0].Name)
			},
		},
		{
			name: "should_leave_another_users_tag_of_the_same_name_alone",
			fn: func(t *testing.T) {
				s.CreateExpense(t, user.ID, newExpenseParams(
					category.ID, "retag mine", 100, 1735689600, []string{"rt_shared_name"}))
				foreign := s.CreateExpense(t, otherUser.ID, newExpenseParams(
					category.ID, "retag theirs", 100, 1735689600, []string{"rt_shared_name"}))

				result, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{"rt_shared_name"},
					To:   "rt_shared_to",
				})
				require.NoError(t, err)
				require.Equal(t, 1, result.Retagged)

				require.Equal(t, []string{"rt_shared_name"}, expenseTags(t, otherUser.ID, foreign.ID))
				require.NotContains(t, userTagNames(t, otherUser.ID), "rt_shared_to")
			},
		},
		{
			name: "should_fail_when_a_source_tag_does_not_exist",
			fn: func(t *testing.T) {
				s.CreateTag(t, user.ID, "rt_unknown_real")

				_, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{"rt_unknown_real", "rt_unknown_typo"},
					To:   "rt_unknown_to",
				})
				require.ErrorIs(t, err, logic.ErrRetagUnknownTag)
				require.NotContains(t, userTagNames(t, user.ID), "rt_unknown_to",
					"a failed retag must not leave its target behind")
			},
		},
		{
			name: "should_fail_when_the_target_is_also_a_source",
			fn: func(t *testing.T) {
				s.CreateTag(t, user.ID, "rt_self")

				_, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{"rt_self"},
					To:   " RT_SELF",
				})
				require.ErrorIs(t, err, logic.ErrRetagSameTag)
			},
		},
		{
			name: "should_fail_validation_without_sources_or_with_a_long_target",
			fn: func(t *testing.T) {
				_, err := s.Store.Retag(ctx, user.ID, logic.RetagParams{From: []string{" "}, To: "rt_valid"})
				require.ErrorIs(t, err, logic.ErrValidationFailed)

				_, err = s.Store.Retag(ctx, user.ID, logic.RetagParams{
					From: []string{"rt_valid_from"},
					To:   "this_tag_name_is_way_too_long",
				})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

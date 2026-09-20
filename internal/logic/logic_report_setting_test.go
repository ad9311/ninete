package logic_test

import (
	"slices"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestReportSettings(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "report_logic_user", "report_logic_user@example.com", "report_password_1")
	other := s.CreateAuthUser(t, "report_other_user", "report_other_user@example.com", "report_password_2")

	tagOne := s.CreateTag(t, user.ID, "rep_lg_tag_one")
	tagTwo := s.CreateTag(t, user.ID, "rep_lg_tag_two")
	foreignTag := s.CreateTag(t, other.ID, "rep_lg_foreign")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_report_defaults_before_anything_is_saved",
			fn: func(t *testing.T) {
				// No stored row is not an error: an unconfigured report is a
				// valid one, printing a single flat list.
				setting, err := s.Store.FindReportSetting(ctx, other.ID)
				require.NoError(t, err)
				require.Empty(t, setting.TagIDs)
			},
		},
		{
			name: "should_save_and_read_back_the_tags",
			fn: func(t *testing.T) {
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{tagOne.ID, tagTwo.ID},
				}))

				setting, err := s.Store.FindReportSetting(ctx, user.ID)
				require.NoError(t, err)
				require.ElementsMatch(t, []int{tagOne.ID, tagTwo.ID}, setting.TagIDs)
			},
		},
		{
			name: "should_replace_the_tag_list_wholesale",
			fn: func(t *testing.T) {
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{tagOne.ID, tagTwo.ID},
				}))
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{tagTwo.ID},
				}))

				setting, err := s.Store.FindReportSetting(ctx, user.ID)
				require.NoError(t, err)
				require.Equal(t, []int{tagTwo.ID}, setting.TagIDs)
			},
		},
		{
			name: "should_accept_an_empty_tag_list",
			fn: func(t *testing.T) {
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{tagOne.ID},
				}))
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: nil,
				}))

				setting, err := s.Store.FindReportSetting(ctx, user.ID)
				require.NoError(t, err)
				require.Empty(t, setting.TagIDs)

				// The row itself survives an empty save — "no grouping tags"
				// is a configuration, not an absence of one.
				_, found, err := s.Queries.SelectReportSettingByUser(ctx, user.ID)
				require.NoError(t, err)
				require.True(t, found)
			},
		},
		{
			name: "should_ignore_a_tag_belonging_to_another_user",
			fn: func(t *testing.T) {
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{tagOne.ID, foreignTag.ID},
				}))

				setting, err := s.Store.FindReportSetting(ctx, user.ID)
				require.NoError(t, err)
				require.Equal(t, []int{tagOne.ID}, setting.TagIDs)
			},
		},
		{
			name: "should_tolerate_a_repeated_tag_id",
			fn: func(t *testing.T) {
				// A duplicate would collide with uq_report_setting_tags_setting_tag
				// and fail the whole save, so dedupeIDs drops it first.
				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{tagOne.ID, tagOne.ID},
				}))

				setting, err := s.Store.FindReportSetting(ctx, user.ID)
				require.NoError(t, err)
				require.Equal(t, []int{tagOne.ID}, setting.TagIDs)
			},
		},
		{
			name: "should_drop_the_tag_from_the_settings_when_the_tag_is_deleted",
			fn: func(t *testing.T) {
				doomed := s.CreateTag(t, user.ID, "rep_lg_doomed")

				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: []int{tagOne.ID, doomed.ID},
				}))

				_, err := s.Store.DeleteTag(ctx, doomed.ID, user.ID)
				require.NoError(t, err)

				setting, err := s.Store.FindReportSetting(ctx, user.ID)
				require.NoError(t, err)
				require.False(t, slices.Contains(setting.TagIDs, doomed.ID))
				require.Contains(t, setting.TagIDs, tagOne.ID)
			},
		},
		{
			name: "should_count_a_repeated_tag_once_against_the_limit",
			fn: func(t *testing.T) {
				tagIDs := make([]int, 0, logic.ReportTagLimit+1)
				for range logic.ReportTagLimit + 1 {
					tagIDs = append(tagIDs, tagOne.ID)
				}

				require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: tagIDs,
				}))
			},
		},
		{
			name: "should_be_cleared_by_deleting_all_user_data",
			fn: func(t *testing.T) {
				wiped := s.CreateAuthUser(
					t, "report_wipe_user", "report_wipe_user@example.com", "report_password_3",
				)
				wipedTag := s.CreateTag(t, wiped.ID, "rep_lg_wipe_tag")

				require.NoError(t, s.Store.SaveReportSetting(ctx, wiped.ID, logic.ReportSettingParams{
					TagIDs: []int{wipedTag.ID},
				}))

				require.NoError(t, s.Store.DeleteAllUserData(ctx, wiped.ID))

				// Checked through the repo rather than the store: the settings
				// row carries no field of its own any more, so an orphan left
				// behind is invisible from FindReportSetting, which would
				// answer "no tags" either way.
				_, found, err := s.Queries.SelectReportSettingByUser(ctx, wiped.ID)
				require.NoError(t, err)
				require.False(t, found, "report settings survived the wipe")

				setting, err := s.Store.FindReportSetting(ctx, wiped.ID)
				require.NoError(t, err)
				require.Empty(t, setting.TagIDs)
			},
		},
		{
			name: "should_reject_more_tags_than_the_limit",
			fn: func(t *testing.T) {
				tagIDs := make([]int, 0, logic.ReportTagLimit+1)
				for i := range logic.ReportTagLimit + 1 {
					tagIDs = append(tagIDs, tagOne.ID+i)
				}

				err := s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
					TagIDs: tagIDs,
				})
				require.ErrorIs(t, err, logic.ErrReportTooManyTags)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

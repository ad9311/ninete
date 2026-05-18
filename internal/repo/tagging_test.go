package repo_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestTagNamesByTargetID(t *testing.T) {
	cases := []struct {
		name string
		in   []repo.TagRow
		want map[int][]string
	}{
		{
			name: "should_return_empty_map_for_no_rows",
			in:   nil,
			want: map[int][]string{},
		},
		{
			name: "should_group_rows_by_target_id",
			in: []repo.TagRow{
				{TargetID: 1, TagName: "alpha"},
				{TargetID: 2, TagName: "beta"},
				{TargetID: 1, TagName: "gamma"},
			},
			want: map[int][]string{
				1: {"alpha", "gamma"},
				2: {"beta"},
			},
		},
		{
			name: "should_sort_tag_names_within_each_group",
			in: []repo.TagRow{
				{TargetID: 7, TagName: "zebra"},
				{TargetID: 7, TagName: "alpha"},
				{TargetID: 7, TagName: "mango"},
			},
			want: map[int][]string{
				7: {"alpha", "mango", "zebra"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := repo.TagNamesByTargetID(tc.in)
			require.Equal(t, tc.want, got)
		})
	}
}

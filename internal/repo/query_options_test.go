package repo_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestSortingBuild(t *testing.T) {
	cases := []struct {
		name    string
		sorting repo.Sorting
		want    string
		wantErr bool
	}{
		{
			name:    "should_return_empty_when_unset",
			sorting: repo.Sorting{},
			want:    "",
		},
		{
			name:    "should_break_ties_on_id",
			sorting: repo.Sorting{Field: "date", Order: "DESC"},
			want:    `ORDER BY "date" DESC, "id" DESC`,
		},
		{
			name:    "should_follow_the_requested_direction_on_the_tiebreaker",
			sorting: repo.Sorting{Field: "name", Order: "ASC"},
			want:    `ORDER BY "name" ASC, "id" ASC`,
		},
		{
			name:    "should_not_repeat_id_when_sorting_by_id",
			sorting: repo.Sorting{Field: "id", Order: "ASC"},
			want:    `ORDER BY "id" ASC`,
		},
		{
			name:    "should_reject_an_invalid_order",
			sorting: repo.Sorting{Field: "date", Order: "SIDEWAYS"},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.sorting.Build()

			if c.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

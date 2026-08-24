package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNextSortOrder(t *testing.T) {
	cases := []struct {
		name                                              string
		currentField, currentOrder, column, columnDefault string
		want                                              string
	}{
		{
			"should_use_column_default_when_not_currently_sorting_by_column",
			"count", "DESC", "date", "ASC", "ASC",
		},
		{
			"should_flip_asc_to_desc_when_currently_sorting_by_column_asc",
			"date", "ASC", "date", "ASC", "DESC",
		},
		{
			"should_flip_desc_to_asc_when_currently_sorting_by_column_desc",
			"date", "DESC", "date", "ASC", "ASC",
		},
		{
			"should_use_count_default_when_current_field_is_date",
			"date", "DESC", "count", "DESC", "DESC",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := nextSortOrder(tc.currentField, tc.currentOrder, tc.column, tc.columnDefault)
			require.Equal(t, tc.want, got)
		})
	}
}

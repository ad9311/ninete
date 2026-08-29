package logic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// snakeFieldName is unexported and decides the JSON names the API publishes, so
// it is pinned here rather than inferred from a handler test.
func TestSnakeFieldName(t *testing.T) {
	cases := []struct {
		name  string
		field string
		want  string
	}{
		{name: "should_lower_a_single_word", field: "Name", want: "name"},
		{name: "should_split_words", field: "SaturatedFatG", want: "saturated_fat_g"},
		{name: "should_keep_a_trailing_unit_as_one_word", field: "ProteinG", want: "protein_g"},
		{name: "should_keep_an_acronym_together", field: "UserID", want: "user_id"},
		{name: "should_handle_a_bare_acronym", field: "ID", want: "id"},
		{name: "should_split_a_leading_acronym", field: "IDNumber", want: "id_number"},
		{name: "should_pass_through_an_empty_field", field: "", want: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, snakeFieldName(c.field))
		})
	}
}

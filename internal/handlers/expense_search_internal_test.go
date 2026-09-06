package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/stretchr/testify/require"
)

// The zone that turns a typed day into these bounds lives on the client
// (web/app/lib/dates.ts, localDayStart/localDayEnd), so what is left to check
// here is that the resolved pair reaches the right column and displaces the
// preset range. See docs/spa-migration.md §3.6 for why the split falls there.
func TestExpenseSearchBounds(t *testing.T) {
	// A billed-date preset range, as apiExpenseListOpts would have layered it.
	presetOpts := func() repo.QueryOptions {
		return repo.QueryOptions{
			Filters: repo.Filters{
				FilterFields: []repo.FilterField{
					{Name: "user_id", Value: 1, Operator: "="},
					{Name: searchFieldBilled, Value: int64(100), Operator: ">="},
					{Name: searchFieldBilled, Value: int64(200), Operator: "<"},
				},
				Connector: "AND",
			},
		}
	}

	parse := func(t *testing.T, query string) expenseSearch {
		t.Helper()
		r := httptest.NewRequest("GET", "/api/expenses?"+query, nil)
		search, err := parseExpenseSearch(r)
		require.NoError(t, err)

		return search
	}

	fieldNames := func(opts repo.QueryOptions) []string {
		names := make([]string, 0, len(opts.Filters.FilterFields))
		for _, f := range opts.Filters.FilterFields {
			names = append(names, f.Name)
		}

		return names
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_filter_created_at_and_drop_the_billed_preset",
			func(t *testing.T) {
				search := parse(t, "created_start=1000&created_end=2000")
				opts := presetOpts()
				search.apply(&opts, 1)

				require.Equal(
					t,
					[]string{"user_id", searchFieldCreated, searchFieldCreated},
					fieldNames(opts),
				)
				require.Equal(t, int64(1000), opts.Filters.FilterFields[1].Value)
				require.Equal(t, ">=", opts.Filters.FilterFields[1].Operator)
				require.Equal(t, int64(2000), opts.Filters.FilterFields[2].Value)
				require.Equal(t, "<", opts.Filters.FilterFields[2].Operator)
			},
		},
		{
			"should_keep_the_billed_preset_when_no_search_is_active",
			func(t *testing.T) {
				search := parse(t, "")
				opts := presetOpts()
				search.apply(&opts, 1)

				require.Equal(
					t,
					[]string{"user_id", searchFieldBilled, searchFieldBilled},
					fieldNames(opts),
				)
			},
		},
		{
			// A text search widens to all time so older matches are not hidden,
			// unless the request picked a range itself.
			"should_drop_the_billed_preset_for_a_text_search_without_a_range",
			func(t *testing.T) {
				search := parse(t, "q=coffee")
				opts := presetOpts()
				search.apply(&opts, 1)

				require.NotContains(t, fieldNames(opts), searchFieldBilled)
			},
		},
		{
			// explicitRange is set by GetAPIExpenses from the presence of the
			// preset bounds, never parsed from the query — the client resolves
			// its named range itself — so the field is set the way the handler
			// sets it. handle_api_expenses_test.go covers the same path through
			// a real request.
			"should_keep_the_billed_preset_for_a_text_search_with_a_range",
			func(t *testing.T) {
				search := parse(t, "q=coffee")
				search.explicitRange = true
				opts := presetOpts()
				search.apply(&opts, 1)

				require.Contains(t, fieldNames(opts), searchFieldBilled)
			},
		},
		{
			"should_reject_a_half_open_or_inverted_bound_pair",
			func(t *testing.T) {
				for _, query := range []string{
					"created_start=1000",
					"created_end=2000",
					"created_start=2000&created_end=1000",
					"created_start=nope&created_end=2000",
				} {
					r := httptest.NewRequest("GET", "/api/expenses?"+query, nil)
					_, err := parseExpenseSearch(r)
					require.ErrorIs(t, err, ErrAPIInvalidDateRange, query)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

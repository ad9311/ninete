package handlers_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

// TestAPIQueryValidation drives the rules decodeQuery enforces through the real
// router, one route per rule kind. Every rejected case is a genuine
// reproduction, checked against the code before the rules existed, where it
// did one of three things:
//   - answered 200 with a default in its place (per_page, page, category_id,
//     expense sort_field=user_id, the stats sort_field, mode, archived) — the
//     silent fallback that let a drifting client go unnoticed;
//   - answered 500 (a lone sort_order, and a recurrent sort_field naming a
//     column the table lacks), since the sort builder's error was unmapped;
//   - answered 422 without naming the key (the bounds, q and tag length),
//     which the "fields" assertion now requires.
func TestAPIQueryValidation(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()
	_, cookies, _ := apiUser(t, s, "api_query_user", "api_query_user@example.com", "api_query_password")

	get := func(t *testing.T, url string) (int, handlers.APIError) {
		t.Helper()

		res, body := doJSON(t, handler, http.MethodGet, url, nil, cookies, "")

		var apiErr handlers.APIError
		if res.StatusCode != http.StatusOK {
			require.NoError(t, json.Unmarshal(body, &apiErr))
		}

		return res.StatusCode, apiErr
	}

	rejected := []struct {
		url   string
		field string
		rule  string
	}{
		{"/api/expenses?per_page=37", "per_page", "oneof"},
		{"/api/expenses?page=abc", "page", "positive"},
		{"/api/expenses?page=0", "page", "positive"},
		{"/api/expenses?category_id=-2", "category_id", "positive"},
		{"/api/expenses?sort_field=user_id&sort_order=ASC", "sort_field", "oneof"},
		{"/api/expenses?sort_order=asc", "sort_order", "oneof"},
		{"/api/expenses?start=yesterday&end=1", "start", "integer"},
		{"/api/expenses?q=" + strings.Repeat("a", 51), "q", "max"},
		{"/api/expenses?tag=" + strings.Repeat("é", 51), "tag", "max"},
		{"/api/expenses/stats?sort_field=amount", "sort_field", "oneof"},
		{"/api/expenses/budgets?start=1&end=2&mode=weeks", "mode", "oneof"},
		{"/api/expenses/budgets?start=1", "end", "required"},
		{"/api/recurrent-expenses?archived=yes", "archived", "boolean"},
		{"/api/recurrent-expenses?sort_field=date", "sort_field", "oneof"},
		{"/api/dashboard?this_start=1&this_end=2&last_start=0", "last_end", "required"},
	}

	accepted := []string{
		"/api/expenses?per_page=100&page=2&category_id=1",
		"/api/expenses?q=" + strings.Repeat("é", 50),
		// Either half of the sort pair alone takes the default's other half.
		"/api/expenses?sort_field=amount",
		"/api/expenses?sort_order=ASC",
		// An empty value is the same as an absent key.
		"/api/expenses?page=&per_page=",
		"/api/expenses/stats?sort_field=category&sort_order=ASC",
		"/api/expenses/budgets?start=1&end=2&mode=months",
		"/api/recurrent-expenses?archived=true&sort_field=period",
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_reject_a_value_that_breaks_its_rule_naming_the_key",
			fn: func(t *testing.T) {
				for _, c := range rejected {
					status, apiErr := get(t, c.url)
					require.Equal(t, http.StatusUnprocessableEntity, status, c.url)
					require.Equal(t, map[string]string{c.field: c.rule}, apiErr.Fields, c.url)
				}
			},
		},
		{
			name: "should_report_every_broken_key_at_once",
			fn: func(t *testing.T) {
				status, apiErr := get(t, "/api/expenses?page=x&per_page=37&sort_order=up")
				require.Equal(t, http.StatusUnprocessableEntity, status)
				require.Equal(t, map[string]string{
					"page": "positive", "per_page": "oneof", "sort_order": "oneof",
				}, apiErr.Fields)
			},
		},
		{
			name: "should_accept_every_value_its_rule_allows",
			fn: func(t *testing.T) {
				for _, url := range accepted {
					status, apiErr := get(t, url)
					require.Equal(t, http.StatusOK, status, "%s: %v", url, apiErr)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

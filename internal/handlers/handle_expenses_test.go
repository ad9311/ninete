package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/expenses", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_expenses_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_list_1", "exp_list_1@example.com", "exp_password_1")
				cookies := s.AuthCookies(t, "exp_list_1@example.com", "exp_password_1")

				req := spec.NewGetRequest("/expenses", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_display_expense_description_in_body",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "exp_list_2", "exp_list_2@example.com", "exp_password_2")
				category := s.CreateCategory(t, "exp_list_cat_1")
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "Visible expense item", 500, time.Now().Unix()))
				cookies := s.AuthCookies(t, "exp_list_2@example.com", "exp_password_2")

				req := spec.NewGetRequest("/expenses", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Visible expense item")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExpensesNew(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/expenses/new", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_new_expense_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_new_1", "exp_new_1@example.com", "exp_password_1")
				cookies := s.AuthCookies(t, "exp_new_1@example.com", "exp_password_1")

				req := spec.NewGetRequest("/expenses/new", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_expenses_with_valid_form",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_post_1", "exp_post_1@example.com", "exp_password_1")
				category := s.CreateCategory(t, "exp_post_cat_1")
				cookies := s.AuthCookies(t, "exp_post_1@example.com", "exp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/expenses/new", cookies)

				form := url.Values{
					"category_id": {fmt.Sprintf("%d", category.ID)},
					"description": {"New test expense"},
					"amount":      {"2500"},
					"date":        {"2026-01-15T00:00:00Z"},
				}
				req := spec.NewPostRequest("/expenses", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_bad_request_with_invalid_form",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_post_2", "exp_post_2@example.com", "exp_password_2")
				cookies := s.AuthCookies(t, "exp_post_2@example.com", "exp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/expenses/new", cookies)

				form := url.Values{
					"category_id": {"0"},
					"description": {""},
					"amount":      {"0"},
					"date":        {""},
				}
				req := spec.NewPostRequest("/expenses", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExpense(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_expense_show_page",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "exp_show_1", "exp_show_1@example.com", "exp_password_1")
				category := s.CreateCategory(t, "exp_show_cat_1")
				expense := s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "Show expense detail", 1200, 1700000000))
				cookies := s.AuthCookies(t, "exp_show_1@example.com", "exp_password_1")

				req := spec.NewGetRequest(fmt.Sprintf("/expenses/%d", expense.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Show expense detail")
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_show_2", "exp_show_2@example.com", "exp_password_2")
				cookies := s.AuthCookies(t, "exp_show_2@example.com", "exp_password_2")

				req := spec.NewGetRequest("/expenses/999999", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExpensesEdit(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_edit_page_for_existing_expense",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "exp_edit_1", "exp_edit_1@example.com", "exp_password_1")
				category := s.CreateCategory(t, "exp_edit_cat_1")
				expense := s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "Edit this expense", 800, 1700000000))
				cookies := s.AuthCookies(t, "exp_edit_1@example.com", "exp_password_1")

				req := spec.NewGetRequest(fmt.Sprintf("/expenses/%d/edit", expense.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_edit_2", "exp_edit_2@example.com", "exp_password_2")
				cookies := s.AuthCookies(t, "exp_edit_2@example.com", "exp_password_2")

				req := spec.NewGetRequest("/expenses/999999/edit", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostExpensesUpdate(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_expenses_after_valid_update",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "exp_update_1", "exp_update_1@example.com", "exp_password_1")
				category := s.CreateCategory(t, "exp_update_cat_1")
				expense := s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "Before update", 500, 1700000000))
				cookies := s.AuthCookies(t, "exp_update_1@example.com", "exp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, fmt.Sprintf("/expenses/%d/edit", expense.ID), cookies)

				form := url.Values{
					"category_id": {fmt.Sprintf("%d", category.ID)},
					"description": {"After update"},
					"amount":      {"3000"},
					"date":        {"2026-02-01T00:00:00Z"},
				}
				req := spec.NewPostRequest(fmt.Sprintf("/expenses/%d", expense.ID), form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_update_2", "exp_update_2@example.com", "exp_password_2")
				cookies := s.AuthCookies(t, "exp_update_2@example.com", "exp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/expenses/new", cookies)

				form := url.Values{
					"category_id": {"1"},
					"description": {"Does not matter"},
					"amount":      {"1000"},
					"date":        {"2026-01-01T00:00:00Z"},
				}
				req := spec.NewPostRequest("/expenses/999999", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostExpensesDelete(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_expenses_after_valid_delete",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "exp_delete_1", "exp_delete_1@example.com", "exp_password_1")
				category := s.CreateCategory(t, "exp_delete_cat_1")
				expense := s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "Delete me", 300, 1700000000))
				cookies := s.AuthCookies(t, "exp_delete_1@example.com", "exp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, fmt.Sprintf("/expenses/%d", expense.ID), cookies)

				req := spec.NewPostRequest(fmt.Sprintf("/expenses/%d/delete", expense.ID), "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_delete_2", "exp_delete_2@example.com", "exp_password_2")
				cookies := s.AuthCookies(t, "exp_delete_2@example.com", "exp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/expenses/new", cookies)

				req := spec.NewPostRequest("/expenses/999999/delete", "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExpensesStats(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/expenses/stats", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_stats_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_stats_1", "exp_stats_1@example.com", "exp_password_1")
				cookies := s.AuthCookies(t, "exp_stats_1@example.com", "exp_password_1")

				req := spec.NewGetRequest("/expenses/stats", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_display_category_total_in_body",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "exp_stats_2", "exp_stats_2@example.com", "exp_password_2")
				category := s.CreateCategory(t, "exp_stats_cat_1")
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "stats expense 1", 5000, time.Now().Unix()))
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "stats expense 2", 3000, time.Now().Unix()))
				cookies := s.AuthCookies(t, "exp_stats_2@example.com", "exp_password_2")

				req := spec.NewGetRequest("/expenses/stats", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "$80.00")
			},
		},
		{
			name: "should_not_show_other_user_totals",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_stats_3", "exp_stats_3@example.com", "exp_password_3")
				otherUser := s.CreateAuthUser(t, "exp_stats_4", "exp_stats_4@example.com", "exp_password_4")
				category := s.CreateCategory(t, "exp_stats_cat_2")
				s.CreateExpense(t, otherUser.ID, newExpenseParams(category.ID, "other user expense", 9999900, 1736467200))
				cookies := s.AuthCookies(t, "exp_stats_3@example.com", "exp_password_3")

				req := spec.NewGetRequest("/expenses/stats", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "$99,999.00")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExpensesSearch(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "exp_search_1", "exp_search_1@example.com", "exp_search_pass_1")
	category := s.CreateCategory(t, "exp_search_cat_1")
	cookies := s.AuthCookies(t, "exp_search_1@example.com", "exp_search_pass_1")

	jan := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC).Unix()
	feb := time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC).Unix()

	taxi := newExpenseParams(category.ID, "Taxi to airport", 1200, jan)
	taxi.Tags = []string{"travel"}
	s.CreateExpense(t, user.ID, taxi)

	groceries := newExpenseParams(category.ID, "Weekly groceries", 5500, feb)
	groceries.Tags = []string{"home"}
	s.CreateExpense(t, user.ID, groceries)

	// date_range=all_time keeps the preset filter out of the way of the searches.
	get := func(t *testing.T, query string) *httptest.ResponseRecorder {
		t.Helper()

		req := spec.NewGetRequest("/expenses?date_range=all_time&"+query, cookies)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_filter_by_description_substring_case_insensitively",
			fn: func(t *testing.T) {
				rec := get(t, "q=TAXI")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Taxi to airport")
				require.NotContains(t, rec.Body.String(), "Weekly groceries")
			},
		},
		{
			name: "should_treat_like_wildcards_as_literal_characters",
			fn: func(t *testing.T) {
				rec := get(t, "q=%25")

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "Taxi to airport")
				require.NotContains(t, rec.Body.String(), "Weekly groceries")
			},
		},
		{
			name: "should_filter_by_tag",
			fn: func(t *testing.T) {
				rec := get(t, "tag=Travel")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Taxi to airport")
				require.NotContains(t, rec.Body.String(), "Weekly groceries")
			},
		},
		{
			name: "should_filter_by_inclusive_date_bounds",
			fn: func(t *testing.T) {
				rec := get(t, "date_from=2026-01-15&date_to=2026-01-15")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Taxi to airport")
				require.NotContains(t, rec.Body.String(), "Weekly groceries")
			},
		},
		{
			name: "should_combine_search_fields",
			fn: func(t *testing.T) {
				rec := get(t, "q=groceries&tag=travel")

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "Taxi to airport")
				require.NotContains(t, rec.Body.String(), "Weekly groceries")
			},
		},
		{
			name: "should_reject_non_iso_date",
			fn: func(t *testing.T) {
				rec := get(t, "date_from=15/01/2026")

				require.Equal(t, http.StatusBadRequest, rec.Code)
				// Assert on the quoted value from the error message; "YYYY-MM-DD"
				// alone would match the input placeholder and title.
				require.Contains(t, rec.Body.String(), `<p class="form-error-text">`)
				require.Contains(t, rec.Body.String(), `&#34;15/01/2026&#34;`)
			},
		},
		{
			name: "should_reject_inverted_date_range",
			fn: func(t *testing.T) {
				rec := get(t, "date_from=2026-02-01&date_to=2026-01-01")

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "should_reject_overlong_search_term",
			fn: func(t *testing.T) {
				rec := get(t, "q="+url.QueryEscape(strings.Repeat("a", 51)))

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "should_reject_unpadded_date_components",
			fn: func(t *testing.T) {
				rec := get(t, "date_from=2026-1-5")

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "should_widen_to_all_time_when_no_date_range_is_given",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/expenses?q=Taxi", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Taxi to airport")
			},
		},
		{
			name: "should_keep_an_explicit_date_range_over_the_search",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/expenses?date_range=this_month&q=Taxi", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "Taxi to airport")
			},
		},
		{
			name: "should_collapse_the_search_panel_when_no_search_is_active",
			fn: func(t *testing.T) {
				rec := get(t, "")

				require.Equal(t, http.StatusOK, rec.Code)
				// The formatter breaks the tag across lines, so match within it.
				require.NotRegexp(t, `search-panel#toggle"\s+open\b`, rec.Body.String())
			},
		},
		{
			name: "should_open_the_search_panel_when_a_search_is_active",
			fn: func(t *testing.T) {
				rec := get(t, "q=taxi")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Regexp(t, `search-panel#toggle"\s+open\b`, rec.Body.String())
			},
		},
		{
			name: "should_keep_search_params_in_sort_links",
			fn: func(t *testing.T) {
				rec := get(t, "q=taxi&tag=travel")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "q=taxi&amp;tag=travel")
			},
		},
		{
			name: "should_not_leak_other_users_tagged_expenses",
			fn: func(t *testing.T) {
				other := s.CreateAuthUser(t, "exp_search_2", "exp_search_2@example.com", "exp_search_pass_2")
				otherExpense := newExpenseParams(category.ID, "Other user trip", 999, jan)
				otherExpense.Tags = []string{"travel"}
				s.CreateExpense(t, other.ID, otherExpense)

				rec := get(t, "tag=travel")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Taxi to airport")
				require.NotContains(t, rec.Body.String(), "Other user trip")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExpensesSearchDateField(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "exp_df_1", "exp_df_1@example.com", "exp_df_pass_1")
	category := s.CreateCategory(t, "exp_df_cat_1")
	cookies := s.AuthCookies(t, "exp_df_1@example.com", "exp_df_pass_1")

	// Billed long ago, but created now — the two columns disagree, which is the
	// whole point of the toggle.
	oldBilled := time.Date(2020, time.June, 10, 0, 0, 0, 0, time.UTC).Unix()
	s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "Backdated receipt", 700, oldBilled))

	today := time.Now().UTC().Format("2006-01-02")

	get := func(t *testing.T, query string) *httptest.ResponseRecorder {
		t.Helper()

		req := spec.NewGetRequest("/expenses?date_range=all_time&"+query, cookies)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_bound_on_the_billed_date_by_default",
			fn: func(t *testing.T) {
				rec := get(t, "date_from=2020-06-10&date_to=2020-06-10")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Backdated receipt")
			},
		},
		{
			name: "should_exclude_a_backdated_expense_when_bounding_on_created_at",
			fn: func(t *testing.T) {
				rec := get(t, "date_from=2020-06-10&date_to=2020-06-10&date_field=created_at")

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "Backdated receipt")
			},
		},
		{
			name: "should_match_a_backdated_expense_by_its_creation_date",
			fn: func(t *testing.T) {
				rec := get(t, "date_from="+today+"&date_to="+today+"&date_field=created_at")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Backdated receipt")
			},
		},
		{
			name: "should_fall_back_to_the_billed_date_for_an_unknown_field",
			fn: func(t *testing.T) {
				rec := get(t, "date_from=2020-06-10&date_to=2020-06-10&date_field=amount")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Backdated receipt")
			},
		},
		{
			name: "should_carry_the_selected_field_into_sort_links",
			fn: func(t *testing.T) {
				rec := get(t, "date_from="+today+"&date_field=created_at")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "date_field=created_at")
			},
		},
		{
			name: "should_keep_the_default_field_out_of_links",
			fn: func(t *testing.T) {
				rec := get(t, "date_from=2020-06-10")

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "date_field=date")
			},
		},
		{
			name: "should_check_the_toggle_when_bounding_on_created_at",
			fn: func(t *testing.T) {
				rec := get(t, "date_field=created_at")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Regexp(t, `name="date_field"[^>]*checked`, rec.Body.String())
			},
		},
		{
			name: "should_leave_the_toggle_unchecked_by_default",
			fn: func(t *testing.T) {
				rec := get(t, "")

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotRegexp(t, `name="date_field"[^>]*checked`, rec.Body.String())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExpensesPerPage(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "exp_pp_1", "exp_pp_1@example.com", "exp_pp_pass_1")
	category := s.CreateCategory(t, "exp_pp_cat_1")
	cookies := s.AuthCookies(t, "exp_pp_1@example.com", "exp_pp_pass_1")

	// 20 expenses, newest first by date, so item 00 is newest and item 19 oldest.
	base := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC).Unix()
	for i := range 20 {
		description := fmt.Sprintf("pp_item_%02d", i)
		s.CreateExpense(t, user.ID, newExpenseParams(category.ID, description, 100, base-int64(i)*secondsPerDayTest))
	}

	get := func(t *testing.T, query string) *httptest.ResponseRecorder {
		t.Helper()

		req := spec.NewGetRequest("/expenses?date_range=all_time&"+query, cookies)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_show_fifteen_rows_by_default",
			fn: func(t *testing.T) {
				rec := get(t, "")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "pp_item_14")
				require.NotContains(t, rec.Body.String(), "pp_item_15")
			},
		},
		{
			name: "should_honour_a_larger_page_size",
			fn: func(t *testing.T) {
				rec := get(t, "per_page=25")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "pp_item_19")
			},
		},
		{
			name: "should_fall_back_to_the_default_for_an_unlisted_page_size",
			fn: func(t *testing.T) {
				rec := get(t, "per_page=100000")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "pp_item_14")
				require.NotContains(t, rec.Body.String(), "pp_item_15")
			},
		},
		{
			name: "should_offer_every_page_size_choice",
			fn: func(t *testing.T) {
				rec := get(t, "")

				require.Equal(t, http.StatusOK, rec.Code)
				// Match on the option label, which is unique to this select —
				// a bare value="15" could collide with a category option.
				for _, choice := range []string{"15", "25", "50", "100"} {
					require.Contains(t, rec.Body.String(), choice+" per page")
				}
			},
		},
		{
			name: "should_honour_the_largest_page_size",
			fn: func(t *testing.T) {
				rec := get(t, "per_page=100")

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "pp_item_19")
				require.Regexp(t, `<option\s+value="100"\s+selected`, rec.Body.String())
			},
		},
		{
			name: "should_mark_the_active_page_size_as_selected",
			fn: func(t *testing.T) {
				rec := get(t, "per_page=50")

				require.Equal(t, http.StatusOK, rec.Code)
				// The formatter breaks option attributes across lines.
				require.Regexp(t, `<option\s+value="50"\s+selected`, rec.Body.String())
			},
		},
		{
			name: "should_render_the_per_page_select_on_a_single_page_listing",
			fn: func(t *testing.T) {
				rec := get(t, "per_page=50")

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), `aria-label="Pagination"`)
				require.Contains(t, rec.Body.String(), `data-filter-target="perPage"`)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

const secondsPerDayTest = int64(24 * 60 * 60)

func newExpenseParams(
	categoryID int,
	description string,
	amount uint64,
	date int64,
) logic.ExpenseParams {
	return logic.ExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  categoryID,
			Description: description,
			Amount:      amount,
		},
		Date: date,
	}
}

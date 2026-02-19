package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "Visible expense item", 500, 1700000000))
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

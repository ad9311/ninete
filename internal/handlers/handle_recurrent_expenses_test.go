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

func TestGetRecurrentExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/recurrent-expenses", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_recurrent_expenses_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_list_1", "rexp_list_1@example.com", "rexp_password_1")
				cookies := s.AuthCookies(t, "rexp_list_1@example.com", "rexp_password_1")

				req := spec.NewGetRequest("/recurrent-expenses", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_display_recurrent_expense_description_in_body",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_list_2", "rexp_list_2@example.com", "rexp_password_2")
				category := s.CreateCategory(t, "rexp_list_cat_1")
				s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Visible recurrent item", 750, 30),
				)
				cookies := s.AuthCookies(t, "rexp_list_2@example.com", "rexp_password_2")

				req := spec.NewGetRequest("/recurrent-expenses", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Visible recurrent item")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetRecurrentExpensesNew(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/recurrent-expenses/new", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_new_recurrent_expense_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_new_1", "rexp_new_1@example.com", "rexp_password_1")
				cookies := s.AuthCookies(t, "rexp_new_1@example.com", "rexp_password_1")

				req := spec.NewGetRequest("/recurrent-expenses/new", cookies)
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

func TestPostRecurrentExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_recurrent_expenses_with_valid_form",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_post_1", "rexp_post_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_post_cat_1")
				cookies := s.AuthCookies(t, "rexp_post_1@example.com", "rexp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := url.Values{
					"category_id": {fmt.Sprintf("%d", category.ID)},
					"description": {"New recurrent expense"},
					"amount":      {"5000"},
					"period":      {"30"},
				}
				req := spec.NewPostRequest("/recurrent-expenses", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/recurrent-expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_bad_request_with_invalid_form",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_post_2", "rexp_post_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_post_2@example.com", "rexp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := url.Values{
					"category_id": {"0"},
					"description": {""},
					"amount":      {"0"},
					"period":      {"0"},
				}
				req := spec.NewPostRequest("/recurrent-expenses", form.Encode(), cookies, csrfToken)
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

func TestGetRecurrentExpense(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_recurrent_expense_show_page",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_show_1", "rexp_show_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_show_cat_1")
				rexp := s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Show recurrent detail", 900, 7),
				)
				cookies := s.AuthCookies(t, "rexp_show_1@example.com", "rexp_password_1")

				req := spec.NewGetRequest(fmt.Sprintf("/recurrent-expenses/%d", rexp.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Show recurrent detail")
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_recurrent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_show_2", "rexp_show_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_show_2@example.com", "rexp_password_2")

				req := spec.NewGetRequest("/recurrent-expenses/999999", cookies)
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

func TestGetRecurrentExpensesEdit(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_edit_page_for_existing_recurrent_expense",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_edit_1", "rexp_edit_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_edit_cat_1")
				rexp := s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Edit this recurrent", 400, 14),
				)
				cookies := s.AuthCookies(t, "rexp_edit_1@example.com", "rexp_password_1")

				req := spec.NewGetRequest(fmt.Sprintf("/recurrent-expenses/%d/edit", rexp.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_recurrent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_edit_2", "rexp_edit_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_edit_2@example.com", "rexp_password_2")

				req := spec.NewGetRequest("/recurrent-expenses/999999/edit", cookies)
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

func TestPostRecurrentExpensesUpdate(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_recurrent_expenses_after_valid_update",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_update_1", "rexp_update_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_update_cat_1")
				rexp := s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Before recurrent update", 600, 7),
				)
				cookies := s.AuthCookies(t, "rexp_update_1@example.com", "rexp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, fmt.Sprintf("/recurrent-expenses/%d/edit", rexp.ID), cookies)

				form := url.Values{
					"category_id": {fmt.Sprintf("%d", category.ID)},
					"description": {"After recurrent update"},
					"amount":      {"7500"},
					"period":      {"14"},
				}
				req := spec.NewPostRequest(fmt.Sprintf("/recurrent-expenses/%d", rexp.ID), form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/recurrent-expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_recurrent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_update_2", "rexp_update_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_update_2@example.com", "rexp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := url.Values{
					"category_id": {"1"},
					"description": {"Does not matter"},
					"amount":      {"1000"},
					"period":      {"7"},
				}
				req := spec.NewPostRequest("/recurrent-expenses/999999", form.Encode(), cookies, csrfToken)
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

func TestPostRecurrentExpensesDelete(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_recurrent_expenses_after_valid_delete",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_delete_1", "rexp_delete_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_delete_cat_1")
				rexp := s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Delete recurrent", 200, 30),
				)
				cookies := s.AuthCookies(t, "rexp_delete_1@example.com", "rexp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, fmt.Sprintf("/recurrent-expenses/%d", rexp.ID), cookies)

				req := spec.NewPostRequest(fmt.Sprintf("/recurrent-expenses/%d/delete", rexp.ID), "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/recurrent-expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_recurrent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_delete_2", "rexp_delete_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_delete_2@example.com", "rexp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				req := spec.NewPostRequest("/recurrent-expenses/999999/delete", "", cookies, csrfToken)
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

func newRecurrentExpenseParams(
	categoryID int,
	description string,
	amount uint64,
	period uint,
) logic.RecurrentExpenseParams {
	return logic.RecurrentExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  categoryID,
			Description: description,
			Amount:      amount,
		},
		Period: period,
	}
}

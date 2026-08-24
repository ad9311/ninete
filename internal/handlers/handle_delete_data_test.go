package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetDeleteData(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/account/delete-data", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_delete_data_page_with_counts",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "acct_del_page", "acct_del_page@example.com", "acct_password_1")
				category := s.CreateCategory(t, "acct del page category")
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "acct del page expense", 500, 1735689600))
				cookies := s.AuthCookies(t, "acct_del_page@example.com", "acct_password_1")

				req := spec.NewGetRequest("/account/delete-data", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "Delete everything")
				require.Contains(t, body, "1 record(s)")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostDeleteDataExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "acct_del_exp", "acct_del_exp@example.com", "acct_password_1")
	category := s.CreateCategory(t, "acct del exp category")
	s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "expense to wipe", 500, 1735689600))
	cookies := s.AuthCookies(t, "acct_del_exp@example.com", "acct_password_1")
	csrfToken, cookies := s.CSRFFrom(t, "/account/delete-data", cookies)

	req := spec.NewPostRequest("/account/expenses/delete-all", "", cookies, csrfToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/account/delete-data", rec.Header().Get("Location"))

	count, err := s.Queries.CountExpensesByUser(t.Context(), user.ID)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestPostDeleteDataAll(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "acct_del_all", "acct_del_all@example.com", "acct_password_1")
	otherUser := s.CreateAuthUser(t, "acct_del_all_other", "acct_del_all_other@example.com", "acct_password_2")
	category := s.CreateCategory(t, "acct del all category")

	taggedParams := newExpenseParams(category.ID, "mine", 500, 1735689600)
	taggedParams.Tags = []string{"acct_wipe_tag"}
	s.CreateExpense(t, user.ID, taggedParams)
	s.CreateExpense(t, otherUser.ID, newExpenseParams(category.ID, "theirs", 600, 1735689600))

	cookies := s.AuthCookies(t, "acct_del_all@example.com", "acct_password_1")
	csrfToken, cookies := s.CSRFFrom(t, "/account/delete-data", cookies)

	req := spec.NewPostRequest("/account/delete-all", "", cookies, csrfToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/account/delete-data", rec.Header().Get("Location"))

	counts, err := s.Store.FindAccountDataCounts(t.Context(), user.ID)
	require.NoError(t, err)
	require.Equal(t, 0, counts.Expenses)
	require.Equal(t, 0, counts.Tags)

	// The other user's data must remain intact.
	otherCount, err := s.Queries.CountExpensesByUser(t.Context(), otherUser.ID)
	require.NoError(t, err)
	require.Equal(t, 1, otherCount)
}

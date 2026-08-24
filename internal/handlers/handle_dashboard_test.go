package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetDashboard(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/dashboard", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_dashboard_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "dash_user_1", "dash_user_1@example.com", "dash_password_1")
				cookies := s.AuthCookies(t, "dash_user_1@example.com", "dash_password_1")

				req := spec.NewGetRequest("/dashboard", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_show_this_month_expense_total",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "dash_user_2", "dash_user_2@example.com", "dash_password_2")
				category := s.CreateCategory(t, "dash_cat_1")
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "Dash expense", 2500, time.Now().Unix()))
				cookies := s.AuthCookies(t, "dash_user_2@example.com", "dash_password_2")

				req := spec.NewGetRequest("/dashboard", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "$25.00")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

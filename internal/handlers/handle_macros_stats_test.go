package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetMacrosStats(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/macros/stats", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_with_default_period_month",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_stats_user_1", "macros_stats_user_1@example.com", "macros_stats_password_1")
				cookies := s.AuthCookies(t, "macros_stats_user_1@example.com", "macros_stats_password_1")

				req := spec.NewGetRequest("/macros/stats", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), `value="month"`)
			},
		},
		{
			name: "should_render_with_week_period_selected",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_stats_user_2", "macros_stats_user_2@example.com", "macros_stats_password_2")
				cookies := s.AuthCookies(t, "macros_stats_user_2@example.com", "macros_stats_password_2")

				req := spec.NewGetRequest("/macros/stats?period=week", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), `value="week"`)
			},
		},
		{
			name: "should_render_with_six_months_period_selected",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_stats_user_3", "macros_stats_user_3@example.com", "macros_stats_password_3")
				cookies := s.AuthCookies(t, "macros_stats_user_3@example.com", "macros_stats_password_3")

				req := spec.NewGetRequest("/macros/stats?period=six_months", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), `value="six_months"`)
			},
		},
		{
			name: "should_include_entry_values_in_chart_data",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "macros_stats_user_4", "macros_stats_user_4@example.com", "macros_stats_password_4")
				y, m, d := time.Now().Date()
				today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix()
				s.CreateMacroEntry(t, user.ID, logic.MacroEntryParams{
					Name:     "Stats test food",
					Kcal:     1234,
					ProteinG: 56,
					CarbsG:   78,
					FatG:     9,
					Date:     today,
					MealType: "other",
				})
				cookies := s.AuthCookies(t, "macros_stats_user_4@example.com", "macros_stats_password_4")

				req := spec.NewGetRequest("/macros/stats?period=week", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "1234")
				require.Contains(t, body, "56")
				require.Contains(t, body, "78")
			},
		},
		{
			name: "should_not_expose_other_users_entries",
			fn: func(t *testing.T) {
				otherUser := s.CreateAuthUser(
					t, "macros_stats_user_5", "macros_stats_user_5@example.com", "macros_stats_password_5",
				)
				y, m, d := time.Now().Date()
				today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix()
				s.CreateMacroEntry(t, otherUser.ID, logic.MacroEntryParams{
					Name:     "Other user food",
					Kcal:     9999,
					ProteinG: 111,
					CarbsG:   222,
					FatG:     333,
					Date:     today,
					MealType: "other",
				})

				s.CreateAuthUser(t, "macros_stats_user_6", "macros_stats_user_6@example.com", "macros_stats_password_6")
				cookies := s.AuthCookies(t, "macros_stats_user_6@example.com", "macros_stats_password_6")

				req := spec.NewGetRequest("/macros/stats?period=week", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "9999")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

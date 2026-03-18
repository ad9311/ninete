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
				require.Contains(t, rec.Body.String(), "Dash expense")
			},
		},
		{
			name: "should_show_no_macro_goals_prompt_when_goals_not_set",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "dash_user_3", "dash_user_3@example.com", "dash_password_3")
				cookies := s.AuthCookies(t, "dash_user_3@example.com", "dash_password_3")

				req := spec.NewGetRequest("/dashboard", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "No macro goals set.")
			},
		},
		{
			name: "should_show_macro_progress_bars_when_goals_are_set",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "dash_user_4", "dash_user_4@example.com", "dash_password_4")
				s.SaveMacroGoal(t, user.ID, logic.MacroGoalParams{
					Kcal:     2000,
					ProteinG: 150,
					CarbsG:   200,
					FatG:     70,
				})
				cookies := s.AuthCookies(t, "dash_user_4@example.com", "dash_password_4")

				req := spec.NewGetRequest("/dashboard", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "2000")
				require.Contains(t, body, "150")
				require.Contains(t, body, "200")
				require.Contains(t, body, "70")
			},
		},
		{
			name: "should_reflect_today_macro_entry_totals_in_progress",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "dash_user_5", "dash_user_5@example.com", "dash_password_5")
				s.SaveMacroGoal(t, user.ID, logic.MacroGoalParams{
					Kcal:     2000,
					ProteinG: 150,
					CarbsG:   200,
					FatG:     70,
				})
				s.CreateMacroEntry(t, user.ID, logic.MacroEntryParams{
					Name:     "Dash lunch",
					Kcal:     600,
					ProteinG: 40,
					CarbsG:   80,
					FatG:     20,
					Date:     time.Now().Unix(),
					MealType: "other",
				})
				cookies := s.AuthCookies(t, "dash_user_5@example.com", "dash_password_5")

				req := spec.NewGetRequest("/dashboard", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "600")
				require.Contains(t, body, "40")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

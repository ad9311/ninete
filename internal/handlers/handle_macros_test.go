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

func TestGetMacros(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/macros", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_index_for_authenticated_user",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_idx_1", "macros_idx_1@example.com", "macros_pw_1")
				cookies := s.AuthCookies(t, "macros_idx_1@example.com", "macros_pw_1")

				req := spec.NewGetRequest("/macros", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_only_show_current_users_entries_for_selected_day",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "macros_idx_2", "macros_idx_2@example.com", "macros_pw_2")
				otherUser := s.CreateAuthUser(t, "macros_idx_3", "macros_idx_3@example.com", "macros_pw_3")
				s.CreateMacroEntry(t, user.ID, newMacroEntryParamsH("ownEntryName", 100, 1772323200))
				s.CreateMacroEntry(t, otherUser.ID, newMacroEntryParamsH("otherEntryName", 200, 1772323200))
				cookies := s.AuthCookies(t, "macros_idx_2@example.com", "macros_pw_2")

				req := spec.NewGetRequest("/macros?date=2026-03-01", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "ownEntryName")
				require.NotContains(t, rec.Body.String(), "otherEntryName")
			},
		},
		{
			name: "should_filter_totals_by_meal_type",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "macros_idx_mt", "macros_idx_mt@example.com", "macros_pw_mt")
				today := int64(1772323200) // 2026-03-01 UTC
				s.CreateMacroEntry(t, user.ID, logic.MacroEntryParams{
					Name: "Morning oats", Kcal: 300, ProteinG: 10, CarbsG: 50, FatG: 5,
					Date: today, MealType: "breakfast",
				})
				s.CreateMacroEntry(t, user.ID, logic.MacroEntryParams{
					Name: "Chicken rice", Kcal: 750, ProteinG: 45, CarbsG: 80, FatG: 15,
					Date: today, MealType: "lunch",
				})
				cookies := s.AuthCookies(t, "macros_idx_mt@example.com", "macros_pw_mt")

				req := spec.NewGetRequest("/macros?date=2026-03-01&meal_type=breakfast", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "Morning oats")
				require.NotContains(t, body, "Chicken rice")
			},
		},
		{
			name: "should_filter_out_entries_from_other_days",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "macros_idx_4", "macros_idx_4@example.com", "macros_pw_4")
				s.CreateMacroEntry(t, user.ID, newMacroEntryParamsH("dayOneEntry", 100, 1772323200)) // 2026-03-01
				s.CreateMacroEntry(t, user.ID, newMacroEntryParamsH("dayTwoEntry", 100, 1772409600)) // 2026-03-02
				cookies := s.AuthCookies(t, "macros_idx_4@example.com", "macros_pw_4")

				req := spec.NewGetRequest("/macros?date=2026-03-01", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "dayOneEntry")
				require.NotContains(t, rec.Body.String(), "dayTwoEntry")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostMacros(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_entry_and_redirect_to_day_view",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_post_1", "macros_post_1@example.com", "macros_pw_1")
				cookies := s.AuthCookies(t, "macros_post_1@example.com", "macros_pw_1")
				csrfToken, cookies := s.CSRFFrom(t, "/macros/new", cookies)

				form := macroEntryFormValues("Chicken breast", "165", "31", "0", "3.6", "2026-03-01T00:00:00Z", "lunch")
				req := spec.NewPostRequest("/macros", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/macros?date=2026-03-01", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_redirect_to_template_creation_when_save_as_template_is_set",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_post_2", "macros_post_2@example.com", "macros_pw_2")
				cookies := s.AuthCookies(t, "macros_post_2@example.com", "macros_pw_2")
				csrfToken, cookies := s.CSRFFrom(t, "/macros/new", cookies)

				form := macroEntryFormValues("Oats", "380", "13", "67", "7", "2026-03-01T00:00:00Z", "breakfast")
				form.Set("save_as_template", "on")
				req := spec.NewPostRequest("/macros", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Contains(t, rec.Header().Get("Location"), "/macros/templates/new?from_entry=")
			},
		},
		{
			name: "should_reject_invalid_meal_type",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_post_3", "macros_post_3@example.com", "macros_pw_3")
				cookies := s.AuthCookies(t, "macros_post_3@example.com", "macros_pw_3")
				csrfToken, cookies := s.CSRFFrom(t, "/macros/new", cookies)

				form := macroEntryFormValues("Snack", "100", "2", "10", "5", "2026-03-01T00:00:00Z", "bogus_meal")
				req := spec.NewPostRequest("/macros", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "should_reject_unparseable_kcal",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_post_4", "macros_post_4@example.com", "macros_pw_4")
				cookies := s.AuthCookies(t, "macros_post_4@example.com", "macros_pw_4")
				csrfToken, cookies := s.CSRFFrom(t, "/macros/new", cookies)

				form := macroEntryFormValues("Chicken", "not_a_number", "31", "0", "3.6", "2026-03-01T00:00:00Z", "lunch")
				req := spec.NewPostRequest("/macros", form.Encode(), cookies, csrfToken)
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

func TestGetMacroEntry(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_show_page_for_owner",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "macros_show_1", "macros_show_1@example.com", "macros_pw_1")
				entry := s.CreateMacroEntry(t, user.ID, newMacroEntryParamsH("Yogurt", 130, 1772323200))
				cookies := s.AuthCookies(t, "macros_show_1@example.com", "macros_pw_1")

				req := spec.NewGetRequest(fmt.Sprintf("/macros/%d", entry.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Yogurt")
			},
		},
		{
			name: "should_return_not_found_for_other_users_entry",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_show_2", "macros_show_2@example.com", "macros_pw_2")
				otherUser := s.CreateAuthUser(t, "macros_show_3", "macros_show_3@example.com", "macros_pw_3")
				entry := s.CreateMacroEntry(t, otherUser.ID, newMacroEntryParamsH("Secret", 100, 1772323200))
				cookies := s.AuthCookies(t, "macros_show_2@example.com", "macros_pw_2")

				req := spec.NewGetRequest(fmt.Sprintf("/macros/%d", entry.ID), cookies)
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

func TestPostMacroEntryUpdate(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_entry_and_redirect_to_day_view",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "macros_upd_1", "macros_upd_1@example.com", "macros_pw_1")
				entry := s.CreateMacroEntry(t, user.ID, newMacroEntryParamsH("Before", 100, 1772323200))
				cookies := s.AuthCookies(t, "macros_upd_1@example.com", "macros_pw_1")
				entryURL := fmt.Sprintf("/macros/%d", entry.ID)
				csrfToken, cookies := s.CSRFFrom(t, entryURL+"/edit", cookies)

				form := macroEntryFormValues("After", "200", "20", "10", "5", "2026-03-02T00:00:00Z", "dinner")
				req := spec.NewPostRequest(entryURL, form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/macros?date=2026-03-02", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_other_users_entry",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_upd_2", "macros_upd_2@example.com", "macros_pw_2")
				otherUser := s.CreateAuthUser(t, "macros_upd_3", "macros_upd_3@example.com", "macros_pw_3")
				entry := s.CreateMacroEntry(t, otherUser.ID, newMacroEntryParamsH("Foreign", 100, 1772323200))
				cookies := s.AuthCookies(t, "macros_upd_2@example.com", "macros_pw_2")
				csrfToken, cookies := s.CSRFFrom(t, "/macros", cookies)

				form := macroEntryFormValues("Hack", "200", "20", "10", "5", "2026-03-02T00:00:00Z", "dinner")
				req := spec.NewPostRequest(
					fmt.Sprintf("/macros/%d", entry.ID),
					form.Encode(), cookies, csrfToken,
				)
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

func TestPostMacroEntryDelete(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_entry_and_redirect_to_index",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "macros_del_1", "macros_del_1@example.com", "macros_pw_1")
				entry := s.CreateMacroEntry(t, user.ID, newMacroEntryParamsH("ToDelete", 100, 1772323200))
				cookies := s.AuthCookies(t, "macros_del_1@example.com", "macros_pw_1")
				entryURL := fmt.Sprintf("/macros/%d", entry.ID)
				csrfToken, cookies := s.CSRFFrom(t, entryURL, cookies)

				req := spec.NewPostRequest(entryURL+"/delete", "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/macros", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_other_users_entry",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_del_2", "macros_del_2@example.com", "macros_pw_2")
				otherUser := s.CreateAuthUser(t, "macros_del_3", "macros_del_3@example.com", "macros_pw_3")
				entry := s.CreateMacroEntry(t, otherUser.ID, newMacroEntryParamsH("Foreign", 100, 1772323200))
				cookies := s.AuthCookies(t, "macros_del_2@example.com", "macros_pw_2")
				csrfToken, cookies := s.CSRFFrom(t, "/macros", cookies)

				req := spec.NewPostRequest(
					fmt.Sprintf("/macros/%d/delete", entry.ID),
					"", cookies, csrfToken,
				)
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

func TestGetMacrosGoals(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_goals_page_with_no_goal_yet",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_goals_1", "macros_goals_1@example.com", "macros_pw_1")
				cookies := s.AuthCookies(t, "macros_goals_1@example.com", "macros_pw_1")

				req := spec.NewGetRequest("/macros/goals", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_render_existing_goal",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "macros_goals_2", "macros_goals_2@example.com", "macros_pw_2")
				s.SaveMacroGoal(t, user.ID, logic.MacroGoalParams{
					Kcal: 2200, ProteinG: 165, CarbsG: 220, FatG: 70,
				})
				cookies := s.AuthCookies(t, "macros_goals_2@example.com", "macros_pw_2")

				req := spec.NewGetRequest("/macros/goals", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "2200")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostMacrosGoals(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_save_goal_and_redirect_to_index",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_pgoals_1", "macros_pgoals_1@example.com", "macros_pw_1")
				cookies := s.AuthCookies(t, "macros_pgoals_1@example.com", "macros_pw_1")
				csrfToken, cookies := s.CSRFFrom(t, "/macros/goals", cookies)

				form := url.Values{
					"kcal":      {"2400"},
					"protein_g": {"180"},
					"carbs_g":   {"240"},
					"fat_g":     {"80"},
				}
				req := spec.NewPostRequest("/macros/goals", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/macros", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_reject_zero_goal_values",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "macros_pgoals_2", "macros_pgoals_2@example.com", "macros_pw_2")
				cookies := s.AuthCookies(t, "macros_pgoals_2@example.com", "macros_pw_2")
				csrfToken, cookies := s.CSRFFrom(t, "/macros/goals", cookies)

				form := url.Values{
					"kcal":      {"0"},
					"protein_g": {"0"},
					"carbs_g":   {"0"},
					"fat_g":     {"0"},
				}
				req := spec.NewPostRequest("/macros/goals", form.Encode(), cookies, csrfToken)
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

func newMacroEntryParamsH(name string, kcal float64, date int64) logic.MacroEntryParams {
	return logic.MacroEntryParams{
		Name:     name,
		Kcal:     kcal,
		ProteinG: 10,
		CarbsG:   20,
		FatG:     5,
		Date:     date,
		MealType: "lunch",
	}
}

func macroEntryFormValues(name, kcal, proteinG, carbsG, fatG, date, mealType string) url.Values {
	return url.Values{
		"name":      {name},
		"kcal":      {kcal},
		"protein_g": {proteinG},
		"carbs_g":   {carbsG},
		"fat_g":     {fatG},
		"date":      {date},
		"meal_type": {mealType},
	}
}

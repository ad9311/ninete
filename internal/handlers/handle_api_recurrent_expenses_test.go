package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

type apiRecurrentExpenseBody struct {
	ID              int      `json:"id"`
	CategoryID      int      `json:"category_id"`
	CategoryName    string   `json:"category_name"`
	Description     string   `json:"description"`
	Amount          uint64   `json:"amount"`
	Period          uint     `json:"period"`
	OccurrenceLimit uint     `json:"occurrence_limit"`
	OccurrenceCount uint     `json:"occurrence_count"`
	Archived        bool     `json:"archived"`
	Tags            []string `json:"tags"`
}

type apiRecurrentExpenseListBody struct {
	Data       []apiRecurrentExpenseBody `json:"data"`
	Pagination struct {
		CurrentPage int    `json:"current_page"`
		TotalPages  int    `json:"total_pages"`
		PerPage     int    `json:"per_page"`
		TotalCount  int    `json:"total_count"`
		HasPrev     bool   `json:"has_prev"`
		HasNext     bool   `json:"has_next"`
		SortField   string `json:"sort_field"`
		SortOrder   string `json:"sort_order"`
		CategoryID  int    `json:"category_id"`
	} `json:"pagination"`
}

// apiUser signs up a fresh authenticated user and mints a CSRF token usable
// against /api/*: the token lives in nosurf's own cookie, shared by both
// chains (§3.2 of docs/spa-migration.md), so an HTML page mints one exactly
// as validly as a future /api/login would.
func apiUser(t *testing.T, s spec.Spec, username, email, password string) (int, []*http.Cookie, string) {
	t.Helper()

	user := s.CreateAuthUser(t, username, email, password)
	cookies := s.AuthCookies(t, email, password)
	csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

	return user.ID, cookies, csrfToken
}

func doJSON(
	t *testing.T,
	handler http.Handler,
	method, url string,
	body any,
	cookies []*http.Cookie,
	csrfToken string,
) (*http.Response, []byte) {
	t.Helper()

	req := spec.NewJSONRequest(method, url, body, cookies, csrfToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, res.Body.Close())
	})

	return res, rec.Body.Bytes()
}

func TestAPIRecurrentExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/api/recurrent-expenses", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusUnauthorized, rec.Code)
				require.Empty(t, rec.Header().Get("Location"))
			},
		},
		{
			name: "should_create_list_show_update_and_delete_a_recurrent_expense",
			fn: func(t *testing.T) {
				_, cookies, csrfToken := apiUser(t, s, "api_re_user_1", "api_re_user_1@example.com", "api_re_password_1")
				category := s.CreateCategory(t, "api_re_cat_1")

				res, body := doJSON(t, handler, http.MethodPost, "/api/recurrent-expenses", map[string]any{
					"category_id":      category.ID,
					"description":      "Streaming",
					"amount":           1999,
					"period":           1,
					"occurrence_limit": 0,
					"tags":             []string{"Media", "media", " Bills "},
				}, cookies, csrfToken)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var created apiRecurrentExpenseBody
				require.NoError(t, json.Unmarshal(body, &created))
				require.NotZero(t, created.ID)
				require.Equal(t, "Streaming", created.Description)
				require.Equal(t, uint64(1999), created.Amount)
				require.False(t, created.Archived)
				// logic.ParseTagNames normalizes: lowercased and deduplicated,
				// so "Media"/"media" collapse to one entry.
				require.ElementsMatch(t, []string{"media", "bills"}, created.Tags)

				res, body = doJSON(t, handler, http.MethodGet, "/api/recurrent-expenses", nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)

				var list apiRecurrentExpenseListBody
				require.NoError(t, json.Unmarshal(body, &list))
				require.Len(t, list.Data, 1)
				require.Equal(t, created.ID, list.Data[0].ID)
				require.Equal(t, category.Name, list.Data[0].CategoryName)
				require.Equal(t, 1, list.Pagination.TotalCount)

				showURL := "/api/recurrent-expenses/" + itoa(created.ID)
				res, body = doJSON(t, handler, http.MethodGet, showURL, nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)

				var shown apiRecurrentExpenseBody
				require.NoError(t, json.Unmarshal(body, &shown))
				require.Equal(t, created.ID, shown.ID)

				res, body = doJSON(t, handler, http.MethodPut, showURL, map[string]any{
					"category_id":      category.ID,
					"description":      "Streaming (annual)",
					"amount":           19999,
					"period":           12,
					"occurrence_limit": 0,
					"tags":             []string{"media"},
				}, cookies, csrfToken)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var updated apiRecurrentExpenseBody
				require.NoError(t, json.Unmarshal(body, &updated))
				require.Equal(t, "Streaming (annual)", updated.Description)
				require.Equal(t, uint(12), updated.Period)
				require.Equal(t, []string{"media"}, updated.Tags)

				res, _ = doJSON(t, handler, http.MethodDelete, showURL, nil, cookies, csrfToken)
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				res, _ = doJSON(t, handler, http.MethodGet, showURL, nil, cookies, "")
				require.Equal(t, http.StatusNotFound, res.StatusCode)
			},
		},
		{
			name: "should_reject_an_invalid_period_with_422_and_field_rules",
			fn: func(t *testing.T) {
				_, cookies, csrfToken := apiUser(t, s, "api_re_user_2", "api_re_user_2@example.com", "api_re_password_2")
				category := s.CreateCategory(t, "api_re_cat_2")

				res, body := doJSON(t, handler, http.MethodPost, "/api/recurrent-expenses", map[string]any{
					"category_id": category.ID,
					"description": "Bad period",
					"amount":      100,
					"period":      0,
				}, cookies, csrfToken)

				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var apiErr handlers.APIError
				require.NoError(t, json.Unmarshal(body, &apiErr))
				// Period is `validate:"required,gt=0"`; a zero value fails
				// "required" before "gt" is ever evaluated.
				require.Equal(t, "required", apiErr.Fields["period"])
			},
		},
		{
			name: "should_not_serve_another_users_recurrent_expense",
			fn: func(t *testing.T) {
				ownerID, _, _ := apiUser(t, s, "api_re_owner", "api_re_owner@example.com", "api_re_password_3")
				category := s.CreateCategory(t, "api_re_cat_3")
				recurrentExpense := s.CreateRecurrentExpense(t, ownerID, logic.RecurrentExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID:  category.ID,
						Description: "Owner only",
						Amount:      500,
					},
					Period: 1,
				})

				_, otherCookies, _ := apiUser(t, s, "api_re_other", "api_re_other@example.com", "api_re_password_4")

				res, _ := doJSON(
					t, handler, http.MethodGet,
					"/api/recurrent-expenses/"+itoa(recurrentExpense.ID),
					nil, otherCookies, "",
				)
				require.Equal(t, http.StatusNotFound, res.StatusCode)
			},
		},
		{
			name: "should_filter_by_archived_and_unarchive",
			fn: func(t *testing.T) {
				ownerID, cookies, csrfToken := apiUser(t, s, "api_re_user_5", "api_re_user_5@example.com", "api_re_password_5")
				category := s.CreateCategory(t, "api_re_cat_5")

				recurrentExpense := s.CreateRecurrentExpense(t, ownerID, logic.RecurrentExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID:  category.ID,
						Description: "Limited",
						Amount:      500,
					},
					Period:          1,
					OccurrenceLimit: 1,
				})
				// A freshly created row has no last_copy_created_at, so it is
				// immediately due. One copy at OccurrenceLimit 1 archives it —
				// the same path the cron job drives in production.
				_, err := s.Store.CopyDueRecurrentExpenses(t.Context(), time.Now())
				require.NoError(t, err)

				res, body := doJSON(
					t, handler, http.MethodGet, "/api/recurrent-expenses?archived=true", nil, cookies, "",
				)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var archivedList apiRecurrentExpenseListBody
				require.NoError(t, json.Unmarshal(body, &archivedList))
				require.Len(t, archivedList.Data, 1)
				require.True(t, archivedList.Data[0].Archived)

				res, body = doJSON(
					t, handler, http.MethodGet, "/api/recurrent-expenses?archived=false", nil, cookies, "",
				)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var activeList apiRecurrentExpenseListBody
				require.NoError(t, json.Unmarshal(body, &activeList))
				require.Empty(t, activeList.Data)

				unarchiveURL := "/api/recurrent-expenses/" + itoa(recurrentExpense.ID) + "/unarchive"
				res, body = doJSON(t, handler, http.MethodPost, unarchiveURL, nil, cookies, csrfToken)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var unarchived apiRecurrentExpenseBody
				require.NoError(t, json.Unmarshal(body, &unarchived))
				require.False(t, unarchived.Archived)
				require.Equal(t, uint(0), unarchived.OccurrenceCount)
			},
		},
		{
			name: "should_reject_a_write_without_a_csrf_token",
			fn: func(t *testing.T) {
				_, cookies, _ := apiUser(t, s, "api_re_user_6", "api_re_user_6@example.com", "api_re_password_6")
				category := s.CreateCategory(t, "api_re_cat_6")

				res, _ := doJSON(t, handler, http.MethodPost, "/api/recurrent-expenses", map[string]any{
					"category_id": category.ID,
					"description": "No token",
					"amount":      100,
					"period":      1,
				}, cookies, "")

				require.Equal(t, http.StatusForbidden, res.StatusCode)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func itoa(id int) string {
	return strconv.Itoa(id)
}

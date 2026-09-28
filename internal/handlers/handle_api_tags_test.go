package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

type apiRetagBody struct {
	Tag struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tag"`
	Retagged int `json:"retagged"`
}

func TestAPITagsRetag(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "api_retag_user", "api_retag_user@example.com", "api_retag_pass_1")
	category := s.CreateCategory(t, "api_retag_cat")

	_, readSecret := s.CreateAPIToken(t, user.ID, "retag read", logic.APITokenScopeRead)
	_, writeSecret := s.CreateAPIToken(t, user.ID, "retag write", logic.APITokenScopeWrite)

	cookies := s.AuthCookies(t, "api_retag_user@example.com", "api_retag_pass_1")

	do := func(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
		t.Helper()

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	post := func(t *testing.T, payload any) *httptest.ResponseRecorder {
		t.Helper()

		token, withToken := s.CSRFFrom(t, "/", cookies)

		return do(t, spec.NewJSONRequest(http.MethodPost, "/api/tags/retag", payload, withToken, token))
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				// A guest session's CSRF token, so the request reaches the auth
				// check instead of being refused by CSRF first.
				token, guest := s.CSRFFrom(t, "/login", nil)
				rec := do(t, spec.NewJSONRequest(http.MethodPost, "/api/tags/retag",
					map[string]any{"from": []string{"a"}, "to": "b"}, guest, token))
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "should_retag_with_a_session",
			fn: func(t *testing.T) {
				s.CreateExpense(t, user.ID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID: category.ID, Description: "api retag session", Amount: 100,
					},
					Date: 1735689600,
					Tags: []string{"api_rt_from"},
				})

				rec := post(t, map[string]any{"from": []string{"api_rt_from"}, "to": "api_rt_to"})
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

				var body apiRetagBody
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				require.Equal(t, "api_rt_to", body.Tag.Name)
				require.Positive(t, body.Tag.ID)
				require.Equal(t, 1, body.Retagged)
			},
		},
		{
			name: "should_answer_422_for_an_unknown_source",
			fn: func(t *testing.T) {
				rec := post(t, map[string]any{"from": []string{"api_rt_missing"}, "to": "api_rt_any"})
				require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), logic.ErrRetagUnknownTag.Error())
			},
		},
		{
			name: "should_answer_422_when_the_target_is_a_source",
			fn: func(t *testing.T) {
				s.CreateTag(t, user.ID, "api_rt_self")

				rec := post(t, map[string]any{"from": []string{"api_rt_self"}, "to": "api_rt_self"})
				require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), logic.ErrRetagSameTag.Error())
			},
		},
		{
			name: "should_answer_422_with_fields_for_an_empty_body",
			fn: func(t *testing.T) {
				rec := post(t, map[string]any{})
				require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), `"from"`)
				require.Contains(t, rec.Body.String(), `"to"`)
			},
		},
		{
			name: "should_refuse_a_read_token",
			fn: func(t *testing.T) {
				s.CreateTag(t, user.ID, "api_rt_read_from")

				rec := do(t, spec.NewBearerRequest(http.MethodPost, "/api/tags/retag",
					map[string]any{"from": []string{"api_rt_read_from"}, "to": "api_rt_read_to"}, readSecret))
				require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			},
		},
		{
			name: "should_retag_with_a_write_token",
			fn: func(t *testing.T) {
				s.CreateTag(t, user.ID, "api_rt_write_from")

				rec := do(t, spec.NewBearerRequest(http.MethodPost, "/api/tags/retag",
					map[string]any{"from": []string{"api_rt_write_from"}, "to": "api_rt_write_to"}, writeSecret))
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

type apiTagListBody struct {
	Data []struct {
		ID                    int    `json:"id"`
		Name                  string `json:"name"`
		ExpenseCount          int    `json:"expense_count"`
		RecurrentExpenseCount int    `json:"recurrent_expense_count"`
	} `json:"data"`
}

func TestAPITags(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "api_tags_user", "api_tags_user@example.com", "api_tags_pass_1")
	other := s.CreateAuthUser(t, "api_tags_other", "api_tags_other@example.com", "api_tags_pass_2")
	category := s.CreateCategory(t, "api_tags_cat")

	_, readSecret := s.CreateAPIToken(t, user.ID, "tags read", logic.APITokenScopeRead)
	_, writeSecret := s.CreateAPIToken(t, user.ID, "tags write", logic.APITokenScopeWrite)

	cookies := s.AuthCookies(t, "api_tags_user@example.com", "api_tags_pass_1")

	do := func(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
		t.Helper()

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	del := func(t *testing.T, url string) *httptest.ResponseRecorder {
		t.Helper()

		token, withToken := s.CSRFFrom(t, "/", cookies)

		return do(t, spec.NewJSONRequest(http.MethodDelete, url, nil, withToken, token))
	}

	list := func(t *testing.T) map[string]int {
		t.Helper()

		rec := do(t, spec.NewGetRequest("/api/tags", cookies))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var body apiTagListBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

		counts := make(map[string]int, len(body.Data))
		for _, tag := range body.Data {
			counts[tag.Name] = tag.ExpenseCount + tag.RecurrentExpenseCount
		}

		return counts
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_list_the_users_tags_with_counts",
			fn: func(t *testing.T) {
				s.CreateExpense(t, user.ID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID: category.ID, Description: "api tags list", Amount: 100,
					},
					Date: 1735689600,
					Tags: []string{"api_tags_used"},
				})
				s.CreateTag(t, user.ID, "api_tags_unused")
				s.CreateTag(t, other.ID, "api_tags_foreign")

				counts := list(t)
				require.Equal(t, 1, counts["api_tags_used"])
				require.Contains(t, counts, "api_tags_unused")
				require.Zero(t, counts["api_tags_unused"])
				require.NotContains(t, counts, "api_tags_foreign")
			},
		},
		{
			name: "should_delete_one_tag",
			fn: func(t *testing.T) {
				tag := s.CreateTag(t, user.ID, "api_tags_del_one")

				rec := del(t, "/api/tags/"+strconv.Itoa(tag.ID))
				require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
				require.NotContains(t, list(t), "api_tags_del_one")
			},
		},
		{
			name: "should_answer_404_for_another_users_tag",
			fn: func(t *testing.T) {
				foreign := s.CreateTag(t, other.ID, "api_tags_not_mine")

				rec := del(t, "/api/tags/"+strconv.Itoa(foreign.ID))
				require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
			},
		},
		{
			name: "should_delete_every_unused_tag",
			fn: func(t *testing.T) {
				s.CreateTag(t, user.ID, "api_tags_orphan")

				rec := del(t, "/api/tags/unused")
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

				var body struct {
					Deleted int `json:"deleted"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				require.Positive(t, body.Deleted)

				counts := list(t)
				require.NotContains(t, counts, "api_tags_orphan")
				require.Contains(t, counts, "api_tags_used")
			},
		},
		{
			name: "should_list_with_a_read_token",
			fn: func(t *testing.T) {
				rec := do(t, spec.NewBearerRequest(http.MethodGet, "/api/tags", nil, readSecret))
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			},
		},
		{
			name: "should_refuse_deletes_to_a_write_token",
			fn: func(t *testing.T) {
				tag := s.CreateTag(t, user.ID, "api_tags_tok_del")

				for _, url := range []string{"/api/tags/" + strconv.Itoa(tag.ID), "/api/tags/unused"} {
					rec := do(t, spec.NewBearerRequest(http.MethodDelete, url, nil, writeSecret))
					require.Equal(t, http.StatusForbidden, rec.Code, url)
				}

				require.Contains(t, list(t), "api_tags_tok_del")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

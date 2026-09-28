package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

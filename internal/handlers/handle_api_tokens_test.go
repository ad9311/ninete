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

type apiTokensBody struct {
	Data []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Prefix     string `json:"prefix"`
		Scope      string `json:"scope"`
		ExpiresAt  *int64 `json:"expires_at"`
		LastUsedAt *int64 `json:"last_used_at"`
		Expired    bool   `json:"expired"`
	} `json:"data"`
	Limit      int   `json:"limit"`
	ExpiryDays []int `json:"expiry_days"`
}

type apiTokenCreatedBody struct {
	Token struct {
		ID     int    `json:"id"`
		Prefix string `json:"prefix"`
		Scope  string `json:"scope"`
	} `json:"token"`
	Secret string `json:"secret"`
}

// TestAPITokensManagement covers the settings page's endpoints, which are
// reached with a browser session.
func TestAPITokensManagement(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	s.CreateAuthUser(t, "api_tok_mgmt_user", "api_tok_mgmt_user@example.com", "api_tok_mgmt_pass_1")
	other := s.CreateAuthUser(t, "api_tok_mgmt_other", "api_tok_mgmt_other@example.com", "api_tok_mgmt_pass_2")

	cookies := s.AuthCookies(t, "api_tok_mgmt_user@example.com", "api_tok_mgmt_pass_1")

	send := func(t *testing.T, method, url string, payload any) *httptest.ResponseRecorder {
		t.Helper()

		token, withToken := s.CSRFFrom(t, "/account/tokens", cookies)
		req := spec.NewJSONRequest(method, url, payload, withToken, token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	list := func(t *testing.T) apiTokensBody {
		t.Helper()

		rec := send(t, http.MethodGet, "/api/tokens", nil)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var body apiTokensBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

		return body
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, spec.NewGetRequest("/api/tokens", nil))

				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "should_create_a_token_and_show_the_secret_once",
			fn: func(t *testing.T) {
				rec := send(t, http.MethodPost, "/api/tokens", map[string]any{
					"name": "laptop mcp", "scope": "write", "expires_in_days": 90,
				})
				require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

				var created apiTokenCreatedBody
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
				require.NotEmpty(t, created.Secret)
				require.Equal(t, "write", created.Token.Scope)

				body := list(t)
				require.Equal(t, logic.APITokenLimit, body.Limit)
				require.Equal(t, logic.APITokenExpiryDays(), body.ExpiryDays)
				require.NotContains(t, rec.Body.String(), "token_hash")

				found := false
				for _, token := range body.Data {
					if token.ID == created.Token.ID {
						found = true
						require.NotNil(t, token.ExpiresAt)
						require.False(t, token.Expired)
					}
				}

				require.True(t, found)

				listRec := send(t, http.MethodGet, "/api/tokens", nil)
				require.NotContains(t, listRec.Body.String(), created.Secret,
					"the secret must not be listed after creation")
			},
		},
		{
			name: "should_answer_422_on_invalid_params",
			fn: func(t *testing.T) {
				rec := send(t, http.MethodPost, "/api/tokens", map[string]any{
					"name": "", "scope": "admin", "expires_in_days": 7,
				})
				require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), `"scope"`)
				require.Contains(t, rec.Body.String(), `"expires_in_days"`)
			},
		},
		{
			name: "should_revoke_a_token",
			fn: func(t *testing.T) {
				rec := send(t, http.MethodPost, "/api/tokens", map[string]any{"name": "to revoke", "scope": "read"})
				require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

				var created apiTokenCreatedBody
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

				rec = send(t, http.MethodDelete, "/api/tokens/"+strconv.Itoa(created.Token.ID), nil)
				require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

				for _, token := range list(t).Data {
					require.NotEqual(t, created.Token.ID, token.ID)
				}

				rec = send(t, http.MethodDelete, "/api/tokens/"+strconv.Itoa(created.Token.ID), nil)
				require.Equal(t, http.StatusNotFound, rec.Code, "revoking twice must read as not found")
			},
		},
		{
			name: "should_not_revoke_another_users_token",
			fn: func(t *testing.T) {
				foreign, secret := s.CreateAPIToken(t, other.ID, "foreign", logic.APITokenScopeRead)

				rec := send(t, http.MethodDelete, "/api/tokens/"+strconv.Itoa(foreign.ID), nil)
				require.Equal(t, http.StatusNotFound, rec.Code)

				_, err := s.Store.AuthenticateAPIToken(t.Context(), secret)
				require.NoError(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

// TestAPIBearerAuth covers what a token may and may not do on the API chain.
func TestAPIBearerAuth(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "api_bearer_user", "api_bearer_user@example.com", "api_bearer_pass_1")
	category := s.CreateCategory(t, "api_bearer_cat")

	_, readSecret := s.CreateAPIToken(t, user.ID, "bearer read", logic.APITokenScopeRead)
	_, writeSecret := s.CreateAPIToken(t, user.ID, "bearer write", logic.APITokenScopeWrite)

	expense := s.CreateExpense(t, user.ID, logic.ExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  category.ID,
			Description: "Bearer expense",
			Amount:      700,
		},
		Date: 1755993600,
	})
	expenseURL := "/api/expenses/" + strconv.Itoa(expense.ID)

	do := func(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
		t.Helper()

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	newExpense := map[string]any{
		"category_id": category.ID,
		"description": "Created by token",
		"amount":      1250,
		"date":        1755993600,
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_read_with_a_read_token_and_no_session",
			fn: func(t *testing.T) {
				rec := do(t, spec.NewBearerRequest(http.MethodGet, expenseURL, nil, readSecret))
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), "Bearer expense")
			},
		},
		{
			name: "should_refuse_writes_to_a_read_token",
			fn: func(t *testing.T) {
				rec := do(t, spec.NewBearerRequest(http.MethodPost, "/api/expenses", newExpense, readSecret))
				require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), "scope")
			},
		},
		{
			name: "should_write_with_a_write_token_and_no_csrf_token",
			fn: func(t *testing.T) {
				rec := do(t, spec.NewBearerRequest(http.MethodPost, "/api/expenses", newExpense, writeSecret))
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				require.Contains(t, rec.Body.String(), "Created by token")
			},
		},
		{
			name: "should_refuse_delete_to_every_scope",
			fn: func(t *testing.T) {
				for _, secret := range []string{readSecret, writeSecret} {
					rec := do(t, spec.NewBearerRequest(http.MethodDelete, expenseURL, nil, secret))
					require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
				}

				rec := do(t, spec.NewBearerRequest(http.MethodGet, expenseURL, nil, readSecret))
				require.Equal(t, http.StatusOK, rec.Code, "the expense must survive the refused delete")
			},
		},
		{
			name: "should_keep_tokens_out_of_routes_off_the_allowlist",
			fn: func(t *testing.T) {
				refused := []struct{ method, url string }{
					{http.MethodGet, "/api/tokens"},
					{http.MethodPost, "/api/tokens"},
					{http.MethodGet, "/api/delete-data"},
					{http.MethodPost, "/api/login"},
				}

				for _, r := range refused {
					rec := do(t, spec.NewBearerRequest(r.method, r.url, map[string]any{}, writeSecret))
					require.Equal(t, http.StatusForbidden, rec.Code, "%s %s: %s", r.method, r.url, rec.Body.String())
				}
			},
		},
		{
			name: "should_answer_401_for_bad_tokens",
			fn: func(t *testing.T) {
				for _, header := range []string{"Bearer nin_unknown", "Bearer ", "Basic abc", "nin_no_scheme"} {
					req := spec.NewBearerRequest(http.MethodGet, "/api/expenses", nil, "")
					req.Header.Set("Authorization", header)

					rec := do(t, req)
					require.Equal(t, http.StatusUnauthorized, rec.Code, header)
					require.NotEmpty(t, rec.Header().Get("WWW-Authenticate"))
				}
			},
		},
		{
			// A genuine reproduction guard: without the no-fallback rule a
			// junk header plus a live session cookie would pass apiAuth.
			name: "should_not_fall_back_to_the_session_when_the_token_is_bad",
			fn: func(t *testing.T) {
				cookies := s.AuthCookies(t, "api_bearer_user@example.com", "api_bearer_pass_1")

				req := spec.NewGetRequest("/api/expenses", cookies)
				req.Header.Set("Authorization", "Bearer nin_not_a_real_token")

				rec := do(t, req)
				require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
			},
		},
		{
			name: "should_reject_a_revoked_token",
			fn: func(t *testing.T) {
				token, secret := s.CreateAPIToken(t, user.ID, "bearer revoked", logic.APITokenScopeRead)
				require.NoError(t, s.Store.RevokeAPIToken(t.Context(), token.ID, user.ID))

				rec := do(t, spec.NewBearerRequest(http.MethodGet, "/api/expenses", nil, secret))
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

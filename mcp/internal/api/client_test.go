package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete-mcp/internal/api"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	const token = "nin_client_test_token"

	serve := func(t *testing.T, h http.HandlerFunc) *api.Client {
		t.Helper()

		srv := httptest.NewServer(h)
		t.Cleanup(srv.Close)

		return api.New(srv.URL, token, "test")
	}

	writeJSON := func(w http.ResponseWriter, status int, body string) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_send_the_bearer_token_and_decode_json",
			fn: func(t *testing.T) {
				client := serve(t, func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, "Bearer "+token, r.Header.Get("Authorization"))
					require.Equal(t, "/api/categories", r.URL.Path)
					require.Empty(t, r.Header.Get("Cookie"))
					writeJSON(w, http.StatusOK, `{"data":[{"id":1,"name":"Food"}]}`)
				})

				var list api.CategoryList
				require.NoError(t, client.Get(t.Context(), "/categories", nil, &list))
				require.Equal(t, []api.Category{{ID: 1, Name: "Food"}}, list.Data)
			},
		},
		{
			name: "should_map_401_to_a_token_error",
			fn: func(t *testing.T) {
				client := serve(t, func(w http.ResponseWriter, _ *http.Request) {
					writeJSON(w, http.StatusUnauthorized, `{"error":"authentication required"}`)
				})

				err := client.Get(t.Context(), "/session", nil, nil)
				require.ErrorIs(t, err, api.ErrUnauthorized)
			},
		},
		{
			name: "should_carry_validation_fields_on_422",
			fn: func(t *testing.T) {
				client := serve(t, func(w http.ResponseWriter, _ *http.Request) {
					writeJSON(w, http.StatusUnprocessableEntity,
						`{"error":"validation failed","fields":{"description":"min"}}`)
				})

				err := client.Post(t.Context(), "/expenses", map[string]any{}, nil)

				var apiErr *api.Error
				require.True(t, errors.As(err, &apiErr))
				require.Equal(t, http.StatusUnprocessableEntity, apiErr.Status)
				require.Equal(t, "min", apiErr.Fields["description"])
				require.Contains(t, err.Error(), "description: min")
			},
		},
		{
			name: "should_carry_the_servers_message_on_403",
			fn: func(t *testing.T) {
				client := serve(t, func(w http.ResponseWriter, _ *http.Request) {
					writeJSON(w, http.StatusForbidden, `{"error":"this token's scope does not allow this request"}`)
				})

				err := client.Put(t.Context(), "/expenses/1", map[string]any{}, nil)
				require.ErrorContains(t, err, "scope")
			},
		},
		{
			// A redirect means the URL is not the API (typically the login
			// page). Following it would forward the request elsewhere.
			name: "should_not_follow_redirects",
			fn: func(t *testing.T) {
				followed := false

				client := serve(t, func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/login" {
						followed = true
					}

					http.Redirect(w, r, "/login", http.StatusSeeOther)
				})

				err := client.Get(t.Context(), "/session", nil, &struct{}{})
				require.ErrorIs(t, err, api.ErrUnexpected)
				require.False(t, followed)
			},
		},
		{
			name: "should_reject_a_non_json_success",
			fn: func(t *testing.T) {
				client := serve(t, func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "text/html")
					_, _ = w.Write([]byte("<html></html>"))
				})

				err := client.Get(t.Context(), "/session", nil, &struct{}{})
				require.ErrorIs(t, err, api.ErrUnexpected)
			},
		},
		{
			name: "should_accept_204_with_no_body",
			fn: func(t *testing.T) {
				client := serve(t, func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})

				require.NoError(t, client.Put(t.Context(), "/report-settings", map[string]any{}, nil))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetAPICategories(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/api/categories", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "should_list_every_category_regardless_of_owner",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "api_cat_user_1", "api_cat_user_1@example.com", "api_cat_password_1")
				cookies := s.AuthCookies(t, "api_cat_user_1@example.com", "api_cat_password_1")
				category := s.CreateCategory(t, "api_cat_1")

				req := spec.NewGetRequest("/api/categories", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)

				var body struct {
					Data []struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

				found := false
				for _, c := range body.Data {
					if c.ID == category.ID {
						found = true

						require.Equal(t, category.Name, c.Name)
					}
				}
				require.True(t, found, "created category missing from /api/categories")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

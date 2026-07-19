package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestPostExpensesQuick(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_ask_for_category_on_first_use",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "quick_h_1", "quick_h_1@example.com", "quick_password_1")
				s.CreateCategory(t, "quick_h_cat_1")
				cookies := s.AuthCookies(t, "quick_h_1@example.com", "quick_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/expenses/new", cookies)

				form := url.Values{"quick_input": {"Netflix, 15.99, today"}}
				req := spec.NewPostRequest("/expenses/quick", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
				require.Contains(t, rec.Body.String(), "Select a category")
			},
		},
		{
			name: "should_create_and_redirect_when_category_provided",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "quick_h_2", "quick_h_2@example.com", "quick_password_2")
				category := s.CreateCategory(t, "quick_h_cat_2")
				cookies := s.AuthCookies(t, "quick_h_2@example.com", "quick_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/expenses/new", cookies)

				form := url.Values{
					"quick_input": {"Spotify, 9.99, today"},
					"category_id": {fmt.Sprintf("%d", category.ID)},
				}
				req := spec.NewPostRequest("/expenses/quick", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_reuse_remembered_category_without_asking",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "quick_h_3", "quick_h_3@example.com", "quick_password_3")
				category := s.CreateCategory(t, "quick_h_cat_3")
				cookies := s.AuthCookies(t, "quick_h_3@example.com", "quick_password_3")
				csrfToken, cookies := s.CSRFFrom(t, "/expenses/new", cookies)

				first := url.Values{
					"quick_input": {"Rent, 1200, today"},
					"category_id": {fmt.Sprintf("%d", category.ID)},
				}
				req := spec.NewPostRequest("/expenses/quick", first.Encode(), cookies, csrfToken)
				handler.ServeHTTP(httptest.NewRecorder(), req)

				second := url.Values{"quick_input": {"Rent, 1300, today"}}
				req = spec.NewPostRequest("/expenses/quick", second.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_bad_request_on_malformed_input",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "quick_h_4", "quick_h_4@example.com", "quick_password_4")
				cookies := s.AuthCookies(t, "quick_h_4@example.com", "quick_password_4")
				csrfToken, cookies := s.CSRFFrom(t, "/expenses/new", cookies)

				form := url.Values{"quick_input": {"missing amount and date"}}
				req := spec.NewPostRequest("/expenses/quick", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
				require.True(t, strings.Contains(rec.Body.String(), "description, amount, date"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

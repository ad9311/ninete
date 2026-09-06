package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

type apiReportSettingsBody struct {
	Timezone       string `json:"timezone"`
	Configured     bool   `json:"configured"`
	SelectedTagIDs []int  `json:"selected_tag_ids"`
	Tags           []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
}

func TestAPIReportSettings(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "api_report_user", "api_report_user@example.com", "api_report_pass_1")
	other := s.CreateAuthUser(t, "api_report_other", "api_report_other@example.com", "api_report_pass_2")

	tagOne := s.CreateTag(t, user.ID, "api_rep_tag_one")
	tagTwo := s.CreateTag(t, user.ID, "api_rep_tag_two")
	foreignTag := s.CreateTag(t, other.ID, "api_rep_foreign")

	cookies := s.AuthCookies(t, "api_report_user@example.com", "api_report_pass_1")

	getSettings := func(t *testing.T) apiReportSettingsBody {
		t.Helper()

		req := spec.NewGetRequest("/api/report-settings", cookies)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var body apiReportSettingsBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

		return body
	}

	put := func(t *testing.T, payload any) *httptest.ResponseRecorder {
		t.Helper()

		token, withToken := s.CSRFFrom(t, "/account/reports", cookies)
		req := spec.NewJSONRequest(http.MethodPut, "/api/report-settings", payload, withToken, token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/api/report-settings", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "should_answer_defaults_and_the_users_own_tags",
			fn: func(t *testing.T) {
				body := getSettings(t)

				require.Equal(t, "UTC", body.Timezone)
				require.False(t, body.Configured)
				require.Empty(t, body.SelectedTagIDs)

				names := make([]string, 0, len(body.Tags))
				for _, tag := range body.Tags {
					names = append(names, tag.Name)
				}

				require.Contains(t, names, tagOne.Name)
				require.Contains(t, names, tagTwo.Name)
				require.NotContains(t, names, foreignTag.Name,
					"another user's tag reached the settings page")
			},
		},
		{
			name: "should_save_and_read_back_the_settings",
			fn: func(t *testing.T) {
				rec := put(t, map[string]any{
					"timezone": "America/Bogota",
					"tag_ids":  []int{tagTwo.ID},
				})
				require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

				body := getSettings(t)
				require.Equal(t, "America/Bogota", body.Timezone)
				require.True(t, body.Configured)
				require.Equal(t, []int{tagTwo.ID}, body.SelectedTagIDs)
			},
		},
		{
			name: "should_reject_an_unknown_timezone",
			fn: func(t *testing.T) {
				rec := put(t, map[string]any{
					"timezone": "Mars/Olympus_Mons",
					"tag_ids":  []int{},
				})
				require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
				require.Contains(t, rec.Body.String(), "unknown time zone")
			},
		},
		{
			name: "should_reject_a_missing_timezone_with_field_errors",
			fn: func(t *testing.T) {
				rec := put(t, map[string]any{"tag_ids": []int{}})
				require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

				var body struct {
					Fields map[string]string `json:"fields"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				require.Contains(t, body.Fields, "timezone")
			},
		},
		{
			name: "should_silently_drop_another_users_tag",
			fn: func(t *testing.T) {
				rec := put(t, map[string]any{
					"timezone": "UTC",
					"tag_ids":  []int{tagOne.ID, foreignTag.ID},
				})
				require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

				body := getSettings(t)
				require.Equal(t, []int{tagOne.ID}, body.SelectedTagIDs)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

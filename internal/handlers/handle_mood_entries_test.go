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

func TestGetMoodEntries(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/moods", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_moods_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mood_idx_1", "mood_idx_1@example.com", "mood_password_1")
				cookies := s.AuthCookies(t, "mood_idx_1@example.com", "mood_password_1")

				req := spec.NewGetRequest("/moods", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_not_show_other_user_entries",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mood_idx_2", "mood_idx_2@example.com", "mood_password_2")
				otherUser := s.CreateAuthUser(t, "mood_idx_3", "mood_idx_3@example.com", "mood_password_3")
				s.CreateMoodEntry(t, otherUser.ID, newMoodEntryParamsH("Happy", "secret note", 1735689600, nil))
				cookies := s.AuthCookies(t, "mood_idx_2@example.com", "mood_password_2")

				req := spec.NewGetRequest("/moods", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "secret note")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostMoodEntries(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_mood_entry_and_redirect",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mood_post_1", "mood_post_1@example.com", "mood_password_1")
				cookies := s.AuthCookies(t, "mood_post_1@example.com", "mood_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/moods/new", cookies)

				form := moodFormValues("Happy", "great day", "2026-01-15T00:00:00Z", "tag_a, tag_b")
				req := spec.NewPostRequest("/moods", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/moods", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_reject_invalid_mood",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mood_post_2", "mood_post_2@example.com", "mood_password_2")
				cookies := s.AuthCookies(t, "mood_post_2@example.com", "mood_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/moods/new", cookies)

				form := moodFormValues("NotAMood", "", "2026-01-15T00:00:00Z", "")
				req := spec.NewPostRequest("/moods", form.Encode(), cookies, csrfToken)
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

func TestGetMoodEntry(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_show_page_for_owner",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "mood_show_1", "mood_show_1@example.com", "mood_password_1")
				entry := s.CreateMoodEntry(t, user.ID, newMoodEntryParamsH("Calm", "after walk", 1735776000, nil))
				cookies := s.AuthCookies(t, "mood_show_1@example.com", "mood_password_1")

				req := spec.NewGetRequest(fmt.Sprintf("/moods/%d", entry.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "after walk")
			},
		},
		{
			name: "should_return_not_found_for_other_users_entry",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mood_show_2", "mood_show_2@example.com", "mood_password_2")
				otherUser := s.CreateAuthUser(t, "mood_show_3", "mood_show_3@example.com", "mood_password_3")
				entry := s.CreateMoodEntry(t, otherUser.ID, newMoodEntryParamsH("Sad", "", 1735862400, nil))
				cookies := s.AuthCookies(t, "mood_show_2@example.com", "mood_password_2")

				req := spec.NewGetRequest(fmt.Sprintf("/moods/%d", entry.ID), cookies)
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

func TestPostMoodEntriesDelete(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_and_redirect",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "mood_del_1", "mood_del_1@example.com", "mood_password_1")
				entry := s.CreateMoodEntry(t, user.ID, newMoodEntryParamsH("Tired", "", 1735948800, nil))
				cookies := s.AuthCookies(t, "mood_del_1@example.com", "mood_password_1")
				entryURL := fmt.Sprintf("/moods/%d", entry.ID)
				csrfToken, cookies := s.CSRFFrom(t, entryURL, cookies)

				req := spec.NewPostRequest(entryURL+"/delete", "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/moods", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_other_users_entry",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mood_del_2", "mood_del_2@example.com", "mood_password_2")
				otherUser := s.CreateAuthUser(t, "mood_del_3", "mood_del_3@example.com", "mood_password_3")
				entry := s.CreateMoodEntry(t, otherUser.ID, newMoodEntryParamsH("Angry", "", 1736035200, nil))
				cookies := s.AuthCookies(t, "mood_del_2@example.com", "mood_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/moods", cookies)

				req := spec.NewPostRequest(
					fmt.Sprintf("/moods/%d/delete", entry.ID),
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

func TestGetMoodEntriesStats(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_stats_page",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "mood_stats_1", "mood_stats_1@example.com", "mood_password_1")
				s.CreateMoodEntry(t, user.ID, newMoodEntryParamsH("Happy", "", 1735689600, nil))
				s.CreateMoodEntry(t, user.ID, newMoodEntryParamsH("Happy", "", 1735776000, nil))
				s.CreateMoodEntry(t, user.ID, newMoodEntryParamsH("Calm", "", 1735862400, nil))
				cookies := s.AuthCookies(t, "mood_stats_1@example.com", "mood_password_1")

				req := spec.NewGetRequest("/moods/stats", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func newMoodEntryParamsH(mood, notes string, loggedAt int64, tags []string) logic.MoodEntryParams {
	return logic.MoodEntryParams{
		Mood:     mood,
		Notes:    notes,
		LoggedAt: loggedAt,
		Tags:     tags,
	}
}

func moodFormValues(mood, notes, date, tags string) url.Values {
	return url.Values{
		"mood":      {mood},
		"notes":     {notes},
		"logged_at": {date},
		"tags":      {tags},
	}
}

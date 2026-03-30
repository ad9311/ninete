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

func TestGetMacroTemplates(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/macros/templates", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_templates_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mt_list_1", "mt_list_1@example.com", "mt_password_1")
				cookies := s.AuthCookies(t, "mt_list_1@example.com", "mt_password_1")

				req := spec.NewGetRequest("/macros/templates", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "MACRO TEMPLATES")
			},
		},
		{
			name: "should_display_template_name_in_body",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "mt_list_2", "mt_list_2@example.com", "mt_password_2")
				s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("Visible template item"))
				cookies := s.AuthCookies(t, "mt_list_2@example.com", "mt_password_2")

				req := spec.NewGetRequest("/macros/templates", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Visible template item")
			},
		},
		{
			name: "should_not_show_other_user_templates",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mt_list_3", "mt_list_3@example.com", "mt_password_3")
				otherUser := s.CreateAuthUser(t, "mt_list_4", "mt_list_4@example.com", "mt_password_4")
				s.CreateMacroTemplate(t, otherUser.ID, newMacroTemplateParams("Other user template"))
				cookies := s.AuthCookies(t, "mt_list_3@example.com", "mt_password_3")

				req := spec.NewGetRequest("/macros/templates", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "Other user template")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetMacroTemplatesNew(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/macros/templates/new", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_new_template_form",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mt_new_1", "mt_new_1@example.com", "mt_password_1")
				cookies := s.AuthCookies(t, "mt_new_1@example.com", "mt_password_1")

				req := spec.NewGetRequest("/macros/templates/new", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "NEW MACRO TEMPLATE")
			},
		},
		{
			name: "should_prefill_from_macro_entry",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "mt_new_2", "mt_new_2@example.com", "mt_password_2")
				entry := s.CreateMacroEntry(t, user.ID, logic.MacroEntryParams{
					Name: "Prefill source", Kcal: 350, ProteinG: 25, CarbsG: 40, FatG: 10,
					Date: 1742083200, MealType: "lunch",
				})
				cookies := s.AuthCookies(t, "mt_new_2@example.com", "mt_password_2")

				req := spec.NewGetRequest(
					fmt.Sprintf("/macros/templates/new?from_entry=%d", entry.ID), cookies,
				)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "Prefill source")
				require.Contains(t, body, "350")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostMacroTemplates(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_template_and_redirect",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mt_post_1", "mt_post_1@example.com", "mt_password_1")
				cookies := s.AuthCookies(t, "mt_post_1@example.com", "mt_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/macros/templates/new", cookies)

				form := macroTemplateFormValues("Post template", "200", "15", "30", "8", "100", "g")
				req := spec.NewPostRequest("/macros/templates", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Contains(t, rec.Header().Get("Location"), "/macros/templates/")
			},
		},
		{
			name: "should_render_errors_for_invalid_params",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mt_post_2", "mt_post_2@example.com", "mt_password_2")
				cookies := s.AuthCookies(t, "mt_post_2@example.com", "mt_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/macros/templates/new", cookies)

				form := macroTemplateFormValues("", "200", "15", "30", "8", "100", "g")
				req := spec.NewPostRequest("/macros/templates", form.Encode(), cookies, csrfToken)
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

func TestGetMacroTemplate(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_template_show_page",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "mt_show_1", "mt_show_1@example.com", "mt_password_1")
				tmpl := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("Show template"))
				cookies := s.AuthCookies(t, "mt_show_1@example.com", "mt_password_1")

				req := spec.NewGetRequest(
					fmt.Sprintf("/macros/templates/%d", tmpl.ID), cookies,
				)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Show template")
			},
		},
		{
			name: "should_return_not_found_for_other_user_template",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mt_show_2", "mt_show_2@example.com", "mt_password_2")
				otherUser := s.CreateAuthUser(t, "mt_show_3", "mt_show_3@example.com", "mt_password_3")
				tmpl := s.CreateMacroTemplate(t, otherUser.ID, newMacroTemplateParams("Secret template"))
				cookies := s.AuthCookies(t, "mt_show_2@example.com", "mt_password_2")

				req := spec.NewGetRequest(
					fmt.Sprintf("/macros/templates/%d", tmpl.ID), cookies,
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

func TestPostMacroTemplateUpdate(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_template_and_redirect",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "mt_upd_1", "mt_upd_1@example.com", "mt_password_1")
				tmpl := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("Before update"))
				cookies := s.AuthCookies(t, "mt_upd_1@example.com", "mt_password_1")
				tmplURL := fmt.Sprintf("/macros/templates/%d", tmpl.ID)
				csrfToken, cookies := s.CSRFFrom(t, tmplURL+"/edit", cookies)

				form := macroTemplateFormValues("After update", "300", "25", "40", "12", "150", "ml")
				req := spec.NewPostRequest(tmplURL, form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, tmplURL, rec.Header().Get("Location"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostMacroTemplateDelete(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_template_and_redirect",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "mt_del_1", "mt_del_1@example.com", "mt_password_1")
				tmpl := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("To delete"))
				cookies := s.AuthCookies(t, "mt_del_1@example.com", "mt_password_1")
				tmplURL := fmt.Sprintf("/macros/templates/%d", tmpl.ID)
				csrfToken, cookies := s.CSRFFrom(t, tmplURL, cookies)

				req := spec.NewPostRequest(tmplURL+"/delete", "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/macros/templates", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_when_deleting_other_user_template",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "mt_del_2", "mt_del_2@example.com", "mt_password_2")
				otherUser := s.CreateAuthUser(t, "mt_del_3", "mt_del_3@example.com", "mt_password_3")
				tmpl := s.CreateMacroTemplate(t, otherUser.ID, newMacroTemplateParams("Not mine"))
				cookies := s.AuthCookies(t, "mt_del_2@example.com", "mt_password_2")

				// Get CSRF from an accessible page since the template show would 404
				csrfToken, cookies := s.CSRFFrom(t, "/macros/templates", cookies)

				req := spec.NewPostRequest(
					fmt.Sprintf("/macros/templates/%d/delete", tmpl.ID),
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

func newMacroTemplateParams(name string) logic.MacroTemplateParams {
	return logic.MacroTemplateParams{
		Name:       name,
		Kcal:       200,
		ProteinG:   15,
		CarbsG:     30,
		FatG:       8,
		Amount:     100,
		AmountUnit: "g",
	}
}

func macroTemplateFormValues(
	name, kcal, proteinG, carbsG, fatG, amount, amountUnit string,
) url.Values {
	return url.Values{
		"name":        {name},
		"kcal":        {kcal},
		"protein_g":   {proteinG},
		"carbs_g":     {carbsG},
		"fat_g":       {fatG},
		"amount":      {amount},
		"amount_unit": {amountUnit},
	}
}

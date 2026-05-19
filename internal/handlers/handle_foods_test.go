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

func TestGetFoods(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/foods", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_foods_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "food_list_1", "food_list_1@example.com", "food_password_1")
				cookies := s.AuthCookies(t, "food_list_1@example.com", "food_password_1")

				req := spec.NewGetRequest("/foods", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Food directory")
			},
		},
		{
			name: "should_display_food_name_in_body",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "food_list_2", "food_list_2@example.com", "food_password_2")
				s.CreateFood(t, user.ID, newFoodParams("Visible food item"))
				cookies := s.AuthCookies(t, "food_list_2@example.com", "food_password_2")

				req := spec.NewGetRequest("/foods", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Visible food item")
			},
		},
		{
			name: "should_not_show_other_user_foods",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "food_list_3", "food_list_3@example.com", "food_password_3")
				otherUser := s.CreateAuthUser(t, "food_list_4", "food_list_4@example.com", "food_password_4")
				s.CreateFood(t, otherUser.ID, newFoodParams("Other user food"))
				cookies := s.AuthCookies(t, "food_list_3@example.com", "food_password_3")

				req := spec.NewGetRequest("/foods", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "Other user food")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetFoodsNew(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/foods/new", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_new_food_form",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "food_new_1", "food_new_1@example.com", "food_password_1")
				cookies := s.AuthCookies(t, "food_new_1@example.com", "food_password_1")

				req := spec.NewGetRequest("/foods/new", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "New food")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostFoods(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_food_and_redirect",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "food_post_1", "food_post_1@example.com", "food_password_1")
				cookies := s.AuthCookies(t, "food_post_1@example.com", "food_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/foods/new", cookies)

				form := foodFormValues("Chicken breast", "165", "31", "0", "3.6")
				req := spec.NewPostRequest("/foods", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Contains(t, rec.Header().Get("Location"), "/foods/")
			},
		},
		{
			name: "should_render_errors_for_invalid_params",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "food_post_2", "food_post_2@example.com", "food_password_2")
				cookies := s.AuthCookies(t, "food_post_2@example.com", "food_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/foods/new", cookies)

				form := foodFormValues("", "165", "31", "0", "3.6")
				req := spec.NewPostRequest("/foods", form.Encode(), cookies, csrfToken)
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

func TestGetFood(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_food_show_page",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "food_show_1", "food_show_1@example.com", "food_password_1")
				food := s.CreateFood(t, user.ID, newFoodParams("Show food"))
				cookies := s.AuthCookies(t, "food_show_1@example.com", "food_password_1")

				req := spec.NewGetRequest(
					fmt.Sprintf("/foods/%d", food.ID), cookies,
				)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Show food")
			},
		},
		{
			name: "should_return_not_found_for_other_user_food",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "food_show_2", "food_show_2@example.com", "food_password_2")
				otherUser := s.CreateAuthUser(t, "food_show_3", "food_show_3@example.com", "food_password_3")
				food := s.CreateFood(t, otherUser.ID, newFoodParams("Secret food"))
				cookies := s.AuthCookies(t, "food_show_2@example.com", "food_password_2")

				req := spec.NewGetRequest(
					fmt.Sprintf("/foods/%d", food.ID), cookies,
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

func TestPostFoodUpdate(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_food_and_redirect",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "food_upd_1", "food_upd_1@example.com", "food_password_1")
				food := s.CreateFood(t, user.ID, newFoodParams("Before update"))
				cookies := s.AuthCookies(t, "food_upd_1@example.com", "food_password_1")
				foodURL := fmt.Sprintf("/foods/%d", food.ID)
				csrfToken, cookies := s.CSRFFrom(t, foodURL+"/edit", cookies)

				form := foodFormValues("After update", "300", "25", "40", "12")
				req := spec.NewPostRequest(foodURL, form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, foodURL, rec.Header().Get("Location"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostFoodDelete(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_food_and_redirect",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "food_del_1", "food_del_1@example.com", "food_password_1")
				food := s.CreateFood(t, user.ID, newFoodParams("To delete"))
				cookies := s.AuthCookies(t, "food_del_1@example.com", "food_password_1")
				foodURL := fmt.Sprintf("/foods/%d", food.ID)
				csrfToken, cookies := s.CSRFFrom(t, foodURL, cookies)

				req := spec.NewPostRequest(foodURL+"/delete", "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/foods", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_when_deleting_other_user_food",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "food_del_2", "food_del_2@example.com", "food_password_2")
				otherUser := s.CreateAuthUser(t, "food_del_3", "food_del_3@example.com", "food_password_3")
				food := s.CreateFood(t, otherUser.ID, newFoodParams("Not mine"))
				cookies := s.AuthCookies(t, "food_del_2@example.com", "food_password_2")

				csrfToken, cookies := s.CSRFFrom(t, "/foods", cookies)

				req := spec.NewPostRequest(
					fmt.Sprintf("/foods/%d/delete", food.ID),
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

func newFoodParams(name string) logic.FoodParams {
	return logic.FoodParams{
		Name:     name,
		Kcal:     200,
		ProteinG: 15,
		CarbsG:   30,
		FatG:     8,
	}
}

func foodFormValues(name, kcal, proteinG, carbsG, fatG string) url.Values {
	return url.Values{
		"name":      {name},
		"kcal":      {kcal},
		"protein_g": {proteinG},
		"carbs_g":   {carbsG},
		"fat_g":     {fatG},
	}
}

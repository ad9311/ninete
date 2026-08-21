package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetRecurrentExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/recurrent-expenses", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_recurrent_expenses_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_list_1", "rexp_list_1@example.com", "rexp_password_1")
				cookies := s.AuthCookies(t, "rexp_list_1@example.com", "rexp_password_1")

				req := spec.NewGetRequest("/recurrent-expenses", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_display_recurrent_expense_description_in_body",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_list_2", "rexp_list_2@example.com", "rexp_password_2")
				category := s.CreateCategory(t, "rexp_list_cat_1")
				s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Visible recurrent item", 750, 30),
				)
				cookies := s.AuthCookies(t, "rexp_list_2@example.com", "rexp_password_2")

				req := spec.NewGetRequest("/recurrent-expenses", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Visible recurrent item")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetRecurrentExpensesNew(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/recurrent-expenses/new", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_new_recurrent_expense_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_new_1", "rexp_new_1@example.com", "rexp_password_1")
				cookies := s.AuthCookies(t, "rexp_new_1@example.com", "rexp_password_1")

				req := spec.NewGetRequest("/recurrent-expenses/new", cookies)
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

func TestPostRecurrentExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_recurrent_expenses_with_valid_form",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_post_1", "rexp_post_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_post_cat_1")
				cookies := s.AuthCookies(t, "rexp_post_1@example.com", "rexp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := recurrentExpenseFormValues(category.ID, "New recurrent expense", "5000", "30", "")
				req := spec.NewPostRequest("/recurrent-expenses", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/recurrent-expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_bad_request_with_invalid_form",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_post_2", "rexp_post_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_post_2@example.com", "rexp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := recurrentExpenseFormValues(0, "", "0", "0", "")
				req := spec.NewPostRequest("/recurrent-expenses", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "should_attach_tags_from_the_form",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_post_3", "rexp_post_3@example.com", "rexp_password_3")
				category := s.CreateCategory(t, "rexp_post_cat_3")
				cookies := s.AuthCookies(t, "rexp_post_3@example.com", "rexp_password_3")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := recurrentExpenseFormValues(
					category.ID,
					"Tagged recurrent expense",
					"5000",
					"30",
					"Rent; fixed",
				)
				req := spec.NewPostRequest("/recurrent-expenses", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)

				recurrentExpense := findRecurrentExpenseByDescription(t, s, user.ID, "Tagged recurrent expense")
				require.Equal(
					t,
					[]string{"fixed", "rent"},
					recurrentExpenseTagNames(t, s, recurrentExpense.ID, user.ID),
				)
			},
		},
		{
			name: "should_round_trip_the_tags_input_when_the_form_is_invalid",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_post_4", "rexp_post_4@example.com", "rexp_password_4")
				cookies := s.AuthCookies(t, "rexp_post_4@example.com", "rexp_password_4")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := recurrentExpenseFormValues(0, "", "0", "0", "kept_tag")
				req := spec.NewPostRequest("/recurrent-expenses", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
				require.Contains(t, rec.Body.String(), "kept_tag")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetRecurrentExpense(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_recurrent_expense_show_page",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_show_1", "rexp_show_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_show_cat_1")
				rexp := s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Show recurrent detail", 900, 7),
				)
				cookies := s.AuthCookies(t, "rexp_show_1@example.com", "rexp_password_1")

				req := spec.NewGetRequest(fmt.Sprintf("/recurrent-expenses/%d", rexp.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Show recurrent detail")
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_recurrent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_show_2", "rexp_show_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_show_2@example.com", "rexp_password_2")

				req := spec.NewGetRequest("/recurrent-expenses/999999", cookies)
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

func TestGetRecurrentExpensesEdit(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_edit_page_for_existing_recurrent_expense",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_edit_1", "rexp_edit_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_edit_cat_1")
				rexp := s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Edit this recurrent", 400, 14),
				)
				cookies := s.AuthCookies(t, "rexp_edit_1@example.com", "rexp_password_1")

				req := spec.NewGetRequest(fmt.Sprintf("/recurrent-expenses/%d/edit", rexp.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_recurrent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_edit_2", "rexp_edit_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_edit_2@example.com", "rexp_password_2")

				req := spec.NewGetRequest("/recurrent-expenses/999999/edit", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name: "should_prefill_the_tags_input_with_the_current_tags",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_edit_3", "rexp_edit_3@example.com", "rexp_password_3")
				category := s.CreateCategory(t, "rexp_edit_cat_3")
				params := newRecurrentExpenseParams(category.ID, "Prefilled recurrent tags", 400, 14)
				params.Tags = []string{"prefill_tag"}
				rexp := s.CreateRecurrentExpense(t, user.ID, params)
				cookies := s.AuthCookies(t, "rexp_edit_3@example.com", "rexp_password_3")

				req := spec.NewGetRequest(fmt.Sprintf("/recurrent-expenses/%d/edit", rexp.ID), cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "prefill_tag")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostRecurrentExpensesUpdate(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_recurrent_expenses_after_valid_update",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_update_1", "rexp_update_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_update_cat_1")
				rexp := s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Before recurrent update", 600, 7),
				)
				cookies := s.AuthCookies(t, "rexp_update_1@example.com", "rexp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, fmt.Sprintf("/recurrent-expenses/%d/edit", rexp.ID), cookies)

				form := recurrentExpenseFormValues(category.ID, "After recurrent update", "7500", "14", "")
				req := spec.NewPostRequest(fmt.Sprintf("/recurrent-expenses/%d", rexp.ID), form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/recurrent-expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_recurrent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_update_2", "rexp_update_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_update_2@example.com", "rexp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := recurrentExpenseFormValues(1, "Does not matter", "1000", "7", "")
				req := spec.NewPostRequest("/recurrent-expenses/999999", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name: "should_replace_tags_from_the_form",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_update_3", "rexp_update_3@example.com", "rexp_password_3")
				category := s.CreateCategory(t, "rexp_update_cat_3")
				params := newRecurrentExpenseParams(category.ID, "Retagged recurrent expense", 600, 7)
				params.Tags = []string{"old_tag"}
				rexp := s.CreateRecurrentExpense(t, user.ID, params)
				cookies := s.AuthCookies(t, "rexp_update_3@example.com", "rexp_password_3")
				csrfToken, cookies := s.CSRFFrom(t, fmt.Sprintf("/recurrent-expenses/%d/edit", rexp.ID), cookies)

				form := recurrentExpenseFormValues(
					category.ID,
					"Retagged recurrent expense",
					"600",
					"7",
					"new_tag",
				)
				req := spec.NewPostRequest(
					fmt.Sprintf("/recurrent-expenses/%d", rexp.ID),
					form.Encode(),
					cookies,
					csrfToken,
				)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, []string{"new_tag"}, recurrentExpenseTagNames(t, s, rexp.ID, user.ID))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostRecurrentExpensesDelete(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_recurrent_expenses_after_valid_delete",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_delete_1", "rexp_delete_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_delete_cat_1")
				rexp := s.CreateRecurrentExpense(t, user.ID,
					newRecurrentExpenseParams(category.ID, "Delete recurrent", 200, 30),
				)
				cookies := s.AuthCookies(t, "rexp_delete_1@example.com", "rexp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, fmt.Sprintf("/recurrent-expenses/%d", rexp.ID), cookies)

				req := spec.NewPostRequest(fmt.Sprintf("/recurrent-expenses/%d/delete", rexp.ID), "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/recurrent-expenses", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_not_found_for_nonexistent_recurrent_expense",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "rexp_delete_2", "rexp_delete_2@example.com", "rexp_password_2")
				cookies := s.AuthCookies(t, "rexp_delete_2@example.com", "rexp_password_2")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				req := spec.NewPostRequest("/recurrent-expenses/999999/delete", "", cookies, csrfToken)
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

func newRecurrentExpenseParams(
	categoryID int,
	description string,
	amount uint64,
	period uint,
) logic.RecurrentExpenseParams {
	return logic.RecurrentExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  categoryID,
			Description: description,
			Amount:      amount,
		},
		Period: period,
	}
}

func recurrentExpenseFormValues(categoryID int, description, amount, period, tags string) url.Values {
	return url.Values{
		"category_id": {fmt.Sprintf("%d", categoryID)},
		"description": {description},
		"amount":      {amount},
		"period":      {period},
		"tags":        {tags},
	}
}

func findRecurrentExpenseByDescription(
	t *testing.T,
	s spec.Spec,
	userID int,
	description string,
) repo.RecurrentExpense {
	t.Helper()

	recurrentExpenses, err := s.Store.FindRecurrentExpenses(t.Context(), repo.QueryOptions{
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{Name: "user_id", Value: userID, Operator: "="},
				{Name: "description", Value: description, Operator: "="},
			},
			Connector: "AND",
		},
	})
	require.NoError(t, err)
	require.Len(t, recurrentExpenses, 1)

	return recurrentExpenses[0]
}

func recurrentExpenseTagNames(t *testing.T, s spec.Spec, recurrentExpenseID, userID int) []string {
	t.Helper()

	tags, err := s.Store.FindRecurrentExpenseTags(t.Context(), recurrentExpenseID, userID)
	require.NoError(t, err)

	return logic.ExtractTagNames(tags)
}

func TestRecurrentExpenseArchiving(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	archive := func(t *testing.T, userID, categoryID int, description string) repo.RecurrentExpense {
		t.Helper()

		params := newRecurrentExpenseParams(categoryID, description, 5000, 1)
		params.OccurrenceLimit = 1
		recurrentExpense := s.CreateRecurrentExpense(t, userID, params)

		_, err := s.Store.CopyDueRecurrentExpenses(t.Context(), time.Now())
		require.NoError(t, err)

		archived, err := s.Store.FindRecurrentExpense(t.Context(), recurrentExpense.ID, userID)
		require.NoError(t, err)
		require.NotNil(t, archived.ArchivedAt)

		return archived
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_store_the_occurrence_limit_from_the_form",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_arch_1", "rexp_arch_1@example.com", "rexp_password_1")
				category := s.CreateCategory(t, "rexp_arch_cat_1")
				cookies := s.AuthCookies(t, "rexp_arch_1@example.com", "rexp_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				form := recurrentExpenseFormValues(category.ID, "Limited recurrent expense", "5000", "1", "")
				form.Set("occurrence_limit", "3")
				req := spec.NewPostRequest("/recurrent-expenses", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)

				created := findRecurrentExpenseByDescription(t, s, user.ID, "Limited recurrent expense")
				require.Equal(t, uint(3), created.OccurrenceLimit)
			},
		},
		{
			name: "should_keep_archived_rows_out_of_the_main_list",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_arch_2", "rexp_arch_2@example.com", "rexp_password_2")
				category := s.CreateCategory(t, "rexp_arch_cat_2")
				archive(t, user.ID, category.ID, "Archived recurrent item 2")
				cookies := s.AuthCookies(t, "rexp_arch_2@example.com", "rexp_password_2")

				req := spec.NewGetRequest("/recurrent-expenses", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "Archived recurrent item 2")
			},
		},
		{
			name: "should_list_archived_rows_on_the_archived_page",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_arch_3", "rexp_arch_3@example.com", "rexp_password_3")
				category := s.CreateCategory(t, "rexp_arch_cat_3")
				archive(t, user.ID, category.ID, "Archived recurrent item 3")
				s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "Active recurrent item 3", 5000, 1),
				)
				cookies := s.AuthCookies(t, "rexp_arch_3@example.com", "rexp_password_3")

				req := spec.NewGetRequest("/recurrent-expenses/archived", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "Archived recurrent item 3")
				require.NotContains(t, rec.Body.String(), "Active recurrent item 3")
			},
		},
		{
			name: "should_unarchive_and_redirect_to_the_recurrent_expense",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "rexp_arch_4", "rexp_arch_4@example.com", "rexp_password_4")
				category := s.CreateCategory(t, "rexp_arch_cat_4")
				archived := archive(t, user.ID, category.ID, "Archived recurrent item 4")
				cookies := s.AuthCookies(t, "rexp_arch_4@example.com", "rexp_password_4")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				path := fmt.Sprintf("/recurrent-expenses/%d/unarchive", archived.ID)
				req := spec.NewPostRequest(path, "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(
					t,
					fmt.Sprintf("/recurrent-expenses/%d", archived.ID),
					rec.Header().Get("Location"),
				)

				updated, err := s.Store.FindRecurrentExpense(t.Context(), archived.ID, user.ID)
				require.NoError(t, err)
				require.Nil(t, updated.ArchivedAt)
				require.Equal(t, uint(0), updated.OccurrenceCount)
			},
		},
		{
			name: "should_not_unarchive_another_users_recurrent_expense",
			fn: func(t *testing.T) {
				owner := s.CreateAuthUser(t, "rexp_arch_5", "rexp_arch_5@example.com", "rexp_password_5")
				s.CreateAuthUser(t, "rexp_arch_6", "rexp_arch_6@example.com", "rexp_password_6")
				category := s.CreateCategory(t, "rexp_arch_cat_5")
				archived := archive(t, owner.ID, category.ID, "Archived recurrent item 5")

				cookies := s.AuthCookies(t, "rexp_arch_6@example.com", "rexp_password_6")
				csrfToken, cookies := s.CSRFFrom(t, "/recurrent-expenses/new", cookies)

				path := fmt.Sprintf("/recurrent-expenses/%d/unarchive", archived.ID)
				req := spec.NewPostRequest(path, "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusNotFound, rec.Code)

				untouched, err := s.Store.FindRecurrentExpense(t.Context(), archived.ID, owner.ID)
				require.NoError(t, err)
				require.NotNil(t, untouched.ArchivedAt)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

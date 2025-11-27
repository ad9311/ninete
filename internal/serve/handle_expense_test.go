package serve_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestGetExpenses(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "getexpenses",
		Email:                "getexpenses@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Get Expenses Category")
	date := time.Now().UTC().Unix()

	expenseOne := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Coffee",
		Amount:      300,
		Date:        date,
	})
	expenseTwo := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Lunch",
		Amount:      1200,
		Date:        date + 10,
	})

	otherUser := f.User(t, logic.SignUpParams{
		Username:             "getexpensesother",
		Email:                "getexpensesother@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	f.Expense(t, repo.InsertExpenseParams{
		UserID:      otherUser.ID,
		CategoryID:  category.ID,
		Description: "Other user expense",
		Amount:      100,
		Date:        date,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_list_user_expenses",
			func(t *testing.T) {
				opts := repo.QueryOptions{
					Sorting: repo.Sorting{Field: "id", Order: "ASC"},
					Pagination: repo.Pagination{
						PerPage: 10,
						Page:    1,
					},
				}
				optsJSON, err := json.Marshal(opts)
				require.NoError(t, err)

				target := "/expenses?query_options=" + url.QueryEscape(string(optsJSON))

				res, req := f.NewRequest(ctx, http.MethodGet, target, nil)
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)

				var payload struct {
					Data  []repo.Expense `json:"data"`
					Error any            `json:"error"`
					Meta  serve.Meta     `json:"meta"`
				}
				testhelper.UnmarshalBody(t, res, &payload)

				require.Nil(t, payload.Error)
				require.Len(t, payload.Data, 2)
				require.Equal(t, expenseOne.ID, payload.Data[0].ID)
				require.Equal(t, expenseTwo.ID, payload.Data[1].ID)
				require.Equal(t, user.ID, payload.Data[0].UserID)
				require.Equal(t, user.ID, payload.Data[1].UserID)
				require.Equal(t, opts.Pagination.PerPage, payload.Meta.PerPage)
				require.Equal(t, opts.Pagination.Page, payload.Meta.Page)
				require.Equal(t, 2, payload.Meta.Rows)
			},
		},
		{
			"should_fail_invalid_sorting",
			func(t *testing.T) {
				opts := repo.QueryOptions{
					Sorting: repo.Sorting{Field: "unknown", Order: "ASC"},
				}
				optsJSON, err := json.Marshal(opts)
				require.NoError(t, err)

				target := "/expenses?query_options=" + url.QueryEscape(string(optsJSON))

				res, req := f.NewRequest(ctx, http.MethodGet, target, nil)
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var payload testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &payload)
				require.Empty(t, payload.Data)
				require.Contains(t, payload.Error, repo.ErrInvalidField.Error())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "getexpense",
		Email:                "getexpense@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Get Expense Category")
	expense := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Groceries",
		Amount:      5000,
		Date:        time.Now().UTC().Unix(),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_get_expense",
			func(t *testing.T) {
				target := fmt.Sprintf("/expenses/%d", expense.ID)

				res, req := f.NewRequest(ctx, http.MethodGet, target, nil)
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)

				var payload testhelper.Response[repo.Expense]
				testhelper.UnmarshalBody(t, res, &payload)
				require.Nil(t, payload.Error)
				require.Equal(t, expense.ID, payload.Data.ID)
				require.Equal(t, expense.Description, payload.Data.Description)
				require.Equal(t, expense.Amount, payload.Data.Amount)
				require.Equal(t, expense.UserID, payload.Data.UserID)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				target := "/expenses/999999"

				res, req := f.NewRequest(ctx, http.MethodGet, target, nil)
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusNotFound, res.Code)

				var payload testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &payload)
				require.Nil(t, payload.Data)
				require.Contains(t, payload.Error, logic.ErrNotFound.Error())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "postexpense",
		Email:                "postexpense@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Post Expense Category")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_create_expense",
			func(t *testing.T) {
				params := logic.ExpenseParams{
					CategoryID:  category.ID,
					Description: "Rent",
					Amount:      100000,
					Date:        time.Now().UTC().Unix(),
				}

				res, req := f.NewRequest(ctx, http.MethodPost, "/expenses", testhelper.MarshalPayload(t, params))
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)

				var payload testhelper.Response[repo.Expense]
				testhelper.UnmarshalBody(t, res, &payload)
				require.Nil(t, payload.Error)
				require.Equal(t, params.Description, payload.Data.Description)
				require.Equal(t, params.Amount, payload.Data.Amount)
				require.Equal(t, user.ID, payload.Data.UserID)
			},
		},
		{
			"should_fail_validation",
			func(t *testing.T) {
				params := logic.ExpenseParams{}

				res, req := f.NewRequest(ctx, http.MethodPost, "/expenses", testhelper.MarshalPayload(t, params))
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var payload testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &payload)
				require.Nil(t, payload.Data)
				require.Contains(t, payload.Error, logic.ErrValidationFailed.Error())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPutExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "putexpense",
		Email:                "putexpense@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Put Expense Category")

	expense := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Old description",
		Amount:      2000,
		Date:        time.Now().UTC().Unix(),
	})

	params := logic.ExpenseParams{
		CategoryID:  category.ID,
		Description: "Updated description",
		Amount:      3000,
		Date:        time.Now().UTC().Unix(),
	}

	target := fmt.Sprintf("/expenses/%d", expense.ID)
	res, req := f.NewRequest(ctx, http.MethodPut, target, testhelper.MarshalPayload(t, params))
	testhelper.SetAuthHeader(req, token.Value)

	f.Server.Router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var payload testhelper.Response[repo.Expense]
	testhelper.UnmarshalBody(t, res, &payload)
	require.Nil(t, payload.Error)
	require.Equal(t, expense.ID, payload.Data.ID)
	require.Equal(t, params.Description, payload.Data.Description)
	require.Equal(t, params.Amount, payload.Data.Amount)
}

func TestDeleteExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "deleteexpense",
		Email:                "deleteexpense@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Delete Expense Category")

	expense := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "To delete",
		Amount:      500,
		Date:        time.Now().UTC().Unix(),
	})

	target := fmt.Sprintf("/expenses/%d", expense.ID)
	res, req := f.NewRequest(ctx, http.MethodDelete, target, nil)
	testhelper.SetAuthHeader(req, token.Value)

	f.Server.Router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var payload struct {
		Data  map[string]int `json:"data"`
		Error any            `json:"error"`
	}
	testhelper.UnmarshalBody(t, res, &payload)
	require.Nil(t, payload.Error)
	require.Equal(t, expense.ID, payload.Data["id"])

	_, err = f.Store.FindExpense(ctx, expense.ID, user.ID)
	require.ErrorIs(t, err, logic.ErrNotFound)
}

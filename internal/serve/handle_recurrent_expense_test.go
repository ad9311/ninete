package serve_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestGetRecurrentExpenses(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "getrecurrexps",
		Email:                "getrecurrexps@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Get Recurrent Expenses Category")
	recurrentOne := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Monthly rent",
		Amount:      90000,
		Period:      1,
	})
	recurrentTwo := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Internet",
		Amount:      8000,
		Period:      1,
	})

	otherUser := f.User(t, logic.SignUpParams{
		Username:             "getrecurrother",
		Email:                "getrecurrother@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	f.RecurrentExpense(t, otherUser.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Other user recurrent",
		Amount:      1000,
		Period:      1,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_list_user_recurrent_expenses",
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

				target := "/recurrent-expenses?query_options=" + url.QueryEscape(string(optsJSON))

				res, req := f.NewRequest(ctx, http.MethodGet, target, nil)
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)

				var payload struct {
					Data  []repo.RecurrentExpense `json:"data"`
					Error any                     `json:"error"`
					Meta  serve.Meta              `json:"meta"`
				}
				testhelper.UnmarshalBody(t, res, &payload)

				require.Nil(t, payload.Error)
				require.Len(t, payload.Data, 2)
				require.Equal(t, recurrentOne.ID, payload.Data[0].ID)
				require.Equal(t, recurrentTwo.ID, payload.Data[1].ID)
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

				target := "/recurrent-expenses?query_options=" + url.QueryEscape(string(optsJSON))

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

func TestGetRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "getrecurrexp",
		Email:                "getrecurrexp@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Get Recurrent Expense Category")
	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Streaming",
		Amount:      1500,
		Period:      1,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_get_recurrent_expense",
			func(t *testing.T) {
				target := fmt.Sprintf("/recurrent-expenses/%d", recurrent.ID)

				res, req := f.NewRequest(ctx, http.MethodGet, target, nil)
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)

				var payload testhelper.Response[repo.RecurrentExpense]
				testhelper.UnmarshalBody(t, res, &payload)
				require.Nil(t, payload.Error)
				require.Equal(t, recurrent.ID, payload.Data.ID)
				require.Equal(t, recurrent.Description, payload.Data.Description)
				require.Equal(t, recurrent.Amount, payload.Data.Amount)
				require.Equal(t, recurrent.UserID, payload.Data.UserID)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				target := "/recurrent-expenses/999999"

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

func TestPutRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "putrecurrexp",
		Email:                "putrecurrexp@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Put Recurrent Expense Category")

	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Old description",
		Amount:      2000,
		Period:      1,
	})

	params := map[string]any{
		"description": "Updated description",
		"amount":      3000,
		"period":      2,
	}

	target := fmt.Sprintf("/recurrent-expenses/%d", recurrent.ID)
	res, req := f.NewRequest(ctx, http.MethodPut, target, testhelper.MarshalPayload(t, params))
	testhelper.SetAuthHeader(req, token.Value)

	f.Server.Router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var payload testhelper.Response[repo.RecurrentExpense]
	testhelper.UnmarshalBody(t, res, &payload)
	require.Nil(t, payload.Error)
	require.Equal(t, recurrent.ID, payload.Data.ID)
	require.Equal(t, params["description"], payload.Data.Description)
	require.Equal(t, uint64(params["amount"].(int)), payload.Data.Amount)
	require.Equal(t, uint(params["period"].(int)), payload.Data.Period)
}

func TestPostRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "postrecurrexp",
		Email:                "postrecurrexp@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Post Recurrent Expense Category")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_create_recurrent_expense",
			func(t *testing.T) {
				params := logic.RecurrentExpenseParams{
					CategoryID:  category.ID,
					Description: "Rent",
					Amount:      100000,
					Period:      1,
				}

				res, req := f.NewRequest(ctx, http.MethodPost, "/recurrent-expenses", testhelper.MarshalPayload(t, params))
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)

				var payload testhelper.Response[repo.RecurrentExpense]
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
				params := logic.RecurrentExpenseParams{}

				res, req := f.NewRequest(ctx, http.MethodPost, "/recurrent-expenses", testhelper.MarshalPayload(t, params))
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

func TestPatchRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "patchrecurrexp",
		Email:                "patchrecurrexp@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Patch Recurrent Expense Category")

	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Old description",
		Amount:      2000,
		Period:      1,
	})

	params := map[string]any{
		"description": "Patched description",
		"amount":      2200,
		"period":      1,
	}

	target := fmt.Sprintf("/recurrent-expenses/%d", recurrent.ID)
	res, req := f.NewRequest(ctx, http.MethodPatch, target, testhelper.MarshalPayload(t, params))
	testhelper.SetAuthHeader(req, token.Value)

	f.Server.Router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var payload testhelper.Response[repo.RecurrentExpense]
	testhelper.UnmarshalBody(t, res, &payload)
	require.Nil(t, payload.Error)
	require.Equal(t, recurrent.ID, payload.Data.ID)
	require.Equal(t, params["description"], payload.Data.Description)
	require.Equal(t, uint64(params["amount"].(int)), payload.Data.Amount)
	require.Equal(t, uint(params["period"].(int)), payload.Data.Period)
}

func TestDeleteRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	userParams := logic.SignUpParams{
		Username:             "delrecurrexp",
		Email:                "delrecurrexp@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, userParams)
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Delete Recurrent Expense Category")

	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "To delete",
		Amount:      500,
		Period:      1,
	})

	target := fmt.Sprintf("/recurrent-expenses/%d", recurrent.ID)
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
	require.Equal(t, recurrent.ID, payload.Data["id"])

	_, err = f.Store.FindRecurrentExpense(ctx, recurrent.ID, user.ID)
	require.ErrorIs(t, err, logic.ErrNotFound)
}

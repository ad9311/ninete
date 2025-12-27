package serve_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestGetRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "getrecurrentexpense",
		Email:                "getrecurrentexpense@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Get Recurrent Expense Category")
	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Internet",
		Amount:      5000,
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

func TestPostRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "postrecurrentexpense",
		Email:                "postrecurrentexpense@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
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
				require.Equal(t, params.Period, payload.Data.Period)
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

func TestPutRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "putrecurrentexpense",
		Email:                "putrecurrentexpense@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	category := f.Category(t, "Put Recurrent Expense Category")
	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Old recurrent",
		Amount:      1000,
		Period:      1,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_update_recurrent_expense",
			func(t *testing.T) {
				params := map[string]any{
					"description": "Updated recurrent",
					"amount":      2500,
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
				require.Equal(t, uint64(2500), payload.Data.Amount)
				require.Equal(t, uint(2), payload.Data.Period)
			},
		},
		{
			"should_fail_validation",
			func(t *testing.T) {
				params := map[string]any{}

				target := fmt.Sprintf("/recurrent-expenses/%d", recurrent.ID)
				res, req := f.NewRequest(ctx, http.MethodPut, target, testhelper.MarshalPayload(t, params))
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var payload testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &payload)
				require.Nil(t, payload.Data)
				require.Contains(t, payload.Error, logic.ErrValidationFailed.Error())
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				params := map[string]any{
					"description": "Missing recurrent",
					"amount":      1200,
					"period":      1,
				}

				target := "/recurrent-expenses/999999"
				res, req := f.NewRequest(ctx, http.MethodPut, target, testhelper.MarshalPayload(t, params))
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

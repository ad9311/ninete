package serve_test

import (
	"net/http"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestGetCategories(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "handlecategories",
		Email:                "handlecategories@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})

	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	categoryOne := f.Category(t, "Handle Categories One")
	categoryTwo := f.Category(t, "Handle Categories Two")

	expected := map[int]repo.Category{
		categoryOne.ID: categoryOne,
		categoryTwo.ID: categoryTwo,
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_list_categories",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/categories", nil)
				testhelper.SetAuthHeader(req, token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)

				var payload testhelper.Response[[]repo.Category]
				testhelper.UnmarshalBody(t, res, &payload)

				require.Nil(t, payload.Error)
				require.GreaterOrEqual(t, len(payload.Data), len(expected))

				found := make(map[int]repo.Category)
				for _, category := range payload.Data {
					found[category.ID] = category
				}

				for id, expectedCategory := range expected {
					category, ok := found[id]
					require.True(t, ok)
					require.Equal(t, expectedCategory.Name, category.Name)
					require.Equal(t, expectedCategory.UID, category.UID)
				}
			},
		},
		{
			"should_reject_missing_auth",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/categories", nil)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)

				var payload testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &payload)
				require.Nil(t, payload.Data)
				require.Contains(t, payload.Error, "invalid auth credentials")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

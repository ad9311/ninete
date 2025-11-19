package serve_test

import (
	"net/http"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestGetHealthz(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	res, req := f.NewRequest(ctx, http.MethodGet, "/healthz", nil)

	f.Server.Router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestGetReadyz(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	res, req := f.NewRequest(ctx, http.MethodGet, "/readyz", nil)

	f.Server.Router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	var resBody testhelper.Response[logic.AppStats]
	testhelper.UnmarshalBody(t, res, &resBody)
	require.Contains(t, resBody.Data.ENV, "test")
	require.Nil(t, resBody.Error)
}

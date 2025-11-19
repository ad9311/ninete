package testhelper

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
	"github.com/stretchr/testify/require"
)

type Factory struct {
	Store  *logic.Store
	Server *serve.Server
	sqlDB  *sql.DB
}

type Response[T any] struct {
	Data  T
	Error any
}

type FailedResponse struct {
	Data  any
	Error string
}

func NewFactory(t *testing.T) Factory {
	t.Helper()

	var f Factory

	app, err := prog.Load()
	require.NoError(t, err)

	sqlDB, err := db.Open()
	require.NoError(t, err)

	queries := repo.New(app, sqlDB)

	store, err := logic.New(app, queries)
	require.NoError(t, err)

	server, err := serve.New(app, store)
	require.NoError(t, err)

	f.Store = store
	f.Server = server
	f.sqlDB = sqlDB

	return f
}

func (*Factory) NewRequest(
	ctx context.Context,
	method, target string,
	body io.Reader,
) (
	*httptest.ResponseRecorder,
	*http.Request,
) {
	res := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, method, target, body)

	if method == http.MethodPost ||
		method == http.MethodPatch ||
		method == http.MethodPut ||
		method == http.MethodDelete {
		SetJSONHeader(req)
	}

	return res, req
}

func (f *Factory) CloseDB(t *testing.T) {
	t.Helper()

	err := f.sqlDB.Close()
	require.NoError(t, err)
}

func (f *Factory) User(t *testing.T, params logic.SignUpParams) repo.SafeUser {
	t.Helper()

	user, err := f.Store.SignUpUser(t.Context(), params)
	require.NoError(t, err)

	return user
}

func (f *Factory) RefreshToken(t *testing.T, userID int) logic.Token {
	t.Helper()

	refreshToken, err := f.Store.NewRefreshToken(t.Context(), userID)
	require.NoError(t, err)

	return refreshToken
}

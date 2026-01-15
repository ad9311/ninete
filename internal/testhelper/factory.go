package testhelper

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
	"github.com/ad9311/ninete/internal/task"
	"github.com/stretchr/testify/require"
)

type Factory struct {
	Store      *logic.Store
	Server     *serve.Server
	TaskConfig *task.Config
	sqlDB      *sql.DB
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
	if err != nil {
		t.Fatalf("failed to load factory program, %v", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("failed to open factory database, %v", err)
	}

	queries := repo.New(app, sqlDB)

	store, err := logic.New(app, queries)
	if err != nil {
		t.Fatalf("failed to instantiate factory store, %v", err)
	}

	server, err := serve.New(app, store)
	if err != nil {
		t.Fatalf("failed to instantiate factory server, %v", err)
	}

	f.Store = store
	f.Server = server
	f.TaskConfig = &task.Config{
		App:   app,
		SQLDB: sqlDB,
		Store: store,
	}
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
	if err != nil {
		t.Fatalf("failed to create factory user, %v", err)
	}

	return user
}

func (f *Factory) SignInUser(
	t *testing.T,
	ctx context.Context,
	params logic.SessionParams,
) Response[serve.SessionResponse] {
	t.Helper()

	body := MarshalPayload(t, params)
	res, req := f.NewRequest(ctx, http.MethodPost, "/auth/sign-in", body)
	f.Server.Router.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		var payload FailedResponse
		UnmarshalBody(t, res, &payload)
		t.Fatalf("failed to sign in user for test, %s", payload.Error)
	}

	var payload Response[serve.SessionResponse]
	UnmarshalBody(t, res, &payload)

	return payload
}

func (f *Factory) RefreshToken(t *testing.T, userID int) logic.Token {
	t.Helper()

	refreshToken, err := f.Store.NewRefreshToken(t.Context(), userID)
	if err != nil {
		t.Fatalf("failed to create factory refresh token, %v", err)
	}

	return refreshToken
}

func (f *Factory) SetRefreshTokenExpiry(
	t *testing.T,
	tokenValue string,
	expiresAt int64,
) {
	t.Helper()

	hash := logic.HashToken(tokenValue)
	_, err := f.sqlDB.Exec(
		`UPDATE "refresh_tokens" SET "expires_at" = ? WHERE "token_hash" = ?`,
		expiresAt,
		hash,
	)
	require.NoError(t, err)
}

func (f *Factory) Category(t *testing.T, name string) repo.Category {
	t.Helper()

	uid := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	category, err := f.Store.CreateCategory(t.Context(), name, uid)
	if err != nil {
		t.Fatalf("failed to create factory category, %v", err)
	}

	return category
}

func (f *Factory) Expense(t *testing.T, userID int, params logic.ExpenseParams) repo.Expense {
	t.Helper()

	logicParams := logic.ExpenseParams{
		CategoryID:  params.CategoryID,
		Description: params.Description,
		Amount:      params.Amount,
		Date:        params.Date,
	}
	expense, err := f.Store.CreateExpense(t.Context(), userID, logicParams)
	if err != nil {
		t.Fatalf("failed to create factory expense, %v", err)
	}

	return expense
}

func (f *Factory) RecurrentExpense(
	t *testing.T,
	userID int,
	params logic.RecurrentExpenseParams,
) repo.RecurrentExpense {
	t.Helper()

	logicParams := logic.RecurrentExpenseParams{
		CategoryID:  params.CategoryID,
		Description: params.Description,
		Amount:      params.Amount,
		Period:      params.Period,
	}
	recurrent, err := f.Store.CreateRecurrentExpense(t.Context(), userID, logicParams)
	if err != nil {
		t.Fatalf("failed to create factory recurrent expense, %v", err)
	}

	return recurrent
}

package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/db"
	"github.com/ad9311/go-api-base/internal/server"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type factoryServer struct {
	router chi.Router
}

type factoryResponse struct {
	Code string         `json:"code"`
	Data map[string]any `json:"data"`
}

type factoryErrorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

type factoryHTTP struct {
	method      string
	target      string
	accessToken string
	body        io.Reader
}

type factorySession struct {
	User               service.SafeUser
	AccessToken        service.Token
	RefreshTokenCookie *http.Cookie
}

func newFactoryServer(t *testing.T) *factoryServer {
	t.Helper()

	config, err := app.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	pool, err := db.Connect(config)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	store, err := service.New(config, pool)
	if err != nil {
		t.Fatalf("failed to create service store: %v", err)
	}

	server := server.New(config, store)

	return &factoryServer{
		router: server.Router,
	}
}

func newHTTPTest(fh factoryHTTP) (*httptest.ResponseRecorder, *http.Request) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(fh.method, fh.target, fh.body)

	req.Header.Set("Content-Type", "application/json")
	if fh.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+fh.accessToken)
	}

	return res, req
}

func decodeJSONBody(t *testing.T, r *httptest.ResponseRecorder, params any) {
	t.Helper()

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		t.Fatalf("failed to decode JSON body: %v", err)
	}
}

func newRequestBody(t *testing.T, params any) []byte {
	t.Helper()

	body, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("failed to marshall struct: %v", err)
	}

	return body
}

func dataToStruct(t *testing.T, bodyData, structData any) {
	t.Helper()

	bytes := newRequestBody(t, bodyData)

	if err := json.Unmarshal(bytes, &structData); err != nil {
		t.Fatalf("failed to unmarshal struct: %v", err)
	}
}

func signUpUser(t *testing.T, fs *factoryServer, body io.Reader) service.SafeUser {
	t.Helper()

	fh := factoryHTTP{
		method:      http.MethodPost,
		target:      "/auth/sign-up",
		body:        body,
		accessToken: "",
	}

	res, req := newHTTPTest(fh)
	fs.router.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		var resBody factoryErrorResponse
		decodeJSONBody(t, res, &resBody)

		log.Printf("server responded with code: %s, error: %s\n", resBody.Code, resBody.Error)
		t.Fatalf("expected status %v, got %v", http.StatusCreated, res.Code)
	}

	var resBody factoryResponse
	decodeJSONBody(t, res, &resBody)

	var data service.SafeUser
	dataToStruct(t, resBody.Data, &data)

	return data
}

func getRefreshToken(t *testing.T, cookie string) *http.Cookie {
	split := strings.Split(cookie, ";")
	if len(split) < 1 {
		t.Fatalf("failed to retrieve refresh token from cookie")
	}

	split = strings.Split(split[0], "=")
	if len(split) != 2 {
		t.Fatalf("failed to retrieve refresh token from cookie")
	}

	return &http.Cookie{
		Name:  "refresh_token",
		Value: split[1],
	}
}

func newFactorySession(t *testing.T, fs *factoryServer, params service.RegistrationParams) factorySession {
	body := newRequestBody(t, service.RegistrationParams{
		Username:             app.SetDefaultString(params.Username, service.FactoryUsername()),
		Email:                app.SetDefaultString(params.Email, service.FactoryUsername()+"@email.com"),
		Password:             app.SetDefaultString(params.Password, "123456789"),
		PasswordConfirmation: app.SetDefaultString(params.PasswordConfirmation, "123456789"),
	})
	user := signUpUser(t, fs, bytes.NewReader(body))

	body = newRequestBody(t, service.SessionParams{
		Email:    user.Email,
		Password: "123456789",
	})

	res, req := newHTTPTest(factoryHTTP{
		method: http.MethodPost,
		target: "/auth/sign-in",
		body:   bytes.NewReader(body),
	})
	fs.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)

	var resBody factoryResponse
	decodeJSONBody(t, res, &resBody)

	var data server.SessionResponse
	dataToStruct(t, resBody.Data, &data)

	cookieStr := res.Header().Get("Set-Cookie")

	cookie := getRefreshToken(t, cookieStr)

	return factorySession{
		User:               user,
		AccessToken:        data.AccessToken,
		RefreshTokenCookie: cookie,
	}
}

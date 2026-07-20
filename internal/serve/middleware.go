package serve

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
)

func (*Server) WithTimeout(dur time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc

				ctx, cancel = context.WithTimeout(ctx, dur)
				defer cancel()
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	guestRoutes := map[string]bool{
		"/login":    true,
		"/register": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		isSignedIn := s.Session.GetBool(r.Context(), handlers.SessionIsUserSignedIn)

		if guestRoutes[path] {
			if isSignedIn {
				http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

				return
			}

			next.ServeHTTP(w, r)

			return
		}

		if strings.HasPrefix(path, "/static/") || path == cspReportPath {
			next.ServeHTTP(w, r)

			return
		}

		if !isSignedIn {
			http.Redirect(w, r, "/login", http.StatusSeeOther)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) csrf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	// Browsers post CSP reports automatically with no CSRF token.
	csrfHandler.ExemptPath(cspReportPath)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   s.app.IsProduction(),
		SameSite: http.SameSiteLaxMode,
		Name:     "ninete_csrf",
	})

	return csrfHandler
}

func (s *Server) setTmplData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		csrf := nosurf.Token(r)

		var currentUser *logic.User
		isUserSignedIn := s.Session.GetBool(ctx, handlers.SessionIsUserSignedIn)
		id := s.Session.GetInt(ctx, handlers.SessionUserID)
		if isUserSignedIn {
			user, err := s.store.FindUser(ctx, id)
			currentUser = &user
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					currentUser = nil
				} else {
					s.app.Logger.Errorf("failed to find current user %v", err)
				}
			}
		}

		nonce, _ := ctx.Value(handlers.KeyCSPNonce).(string)

		templateMap := map[string]any{
			"csrfToken":      csrf,
			"cspNonce":       nonce,
			"error":          "",
			"isUserSignedIn": isUserSignedIn,
			"currentUser":    currentUser,
		}

		ctx = context.WithValue(ctx, handlers.KeyCurrentUser, currentUser)
		ctx = context.WithValue(ctx, handlers.KeyTemplateData, templateMap)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

const maxRequestBodySize = 1 << 20 // 1 MB

func (*Server) limitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		next.ServeHTTP(w, r)
	})
}

func generateNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("%w: %w", ErrNonceGeneration, err)
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

// cspReportPath receives browser CSP violation reports so a blocked inline
// script/style produces a server-side signal instead of failing silently.
const cspReportPath = "/csp-report"

func buildCSP(nonce string) string {
	return "default-src 'self'; " +
		"script-src 'self' 'nonce-" + nonce + "'; " +
		"style-src 'self' 'nonce-" + nonce + "'; " +
		"img-src 'self' data:; " +
		"font-src 'self'; " +
		"connect-src 'self'; " +
		"object-src 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'; " +
		"frame-ancestors 'none'; " +
		"report-uri " + cspReportPath + "; " +
		"report-to csp"
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if s.app.IsProduction() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Static assets carry no nonce'd markup — skip the per-request nonce
		// generation and CSP on the asset hot path.
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)

			return
		}

		nonce, err := generateNonce()
		if err != nil {
			s.app.Logger.Errorf("%v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		w.Header().Set("Reporting-Endpoints", `csp="`+cspReportPath+`"`)
		w.Header().Set("Content-Security-Policy", buildCSP(nonce))
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		ctx := context.WithValue(r.Context(), handlers.KeyCSPNonce, nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) setUpMiddlewares() {
	if !s.app.IsTest() {
		s.Router.Use(middleware.Logger)
	}

	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.RequestID)
	s.Router.Use(s.securityHeaders)
	s.Router.Use(s.limitRequestBody)

	s.Router.Use(s.WithTimeout(5 * time.Second))

	s.Router.Use(s.csrf)

	s.Router.Use(s.setTmplData)
	s.Router.Use(s.AuthMiddleware)

	s.Router.NotFound(s.handlers.NotFound)
	s.Router.MethodNotAllowed(s.handlers.MethodNotAllowed)
}

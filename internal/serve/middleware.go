package serve

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
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

		if path == cspReportPath {
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
			"version":        prog.Version,
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

// baseSecurityHeaders carries the headers every response needs, static assets
// included.
func (s *Server) baseSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if s.app.IsProduction() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// contentSecurityPolicy issues the per-request nonce and the CSP that references
// it. Only rendered pages carry nonce'd markup, so static assets never reach it.
func (s *Server) contentSecurityPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

// Credential attempts allowed per client per window. A person logging in never
// approaches this; a script trying passwords hits it on the first burst.
const (
	authAttemptLimit  = 10
	authAttemptWindow = time.Minute
)

// authRateLimit throttles the credential-checking routes. It keys off
// RemoteAddr, which realClientIP has already rewritten from the proxy's
// forwarded header, so the key is the actual client rather than Caddy.
//
// The returned middleware is shared by every route it guards, so /login and
// /register draw on one budget per client rather than one each.
//
// Disabled under ENV=test: the suite performs a hundred logins from one
// synthetic address, which is exactly the pattern this blocks. The middleware
// itself is covered directly in middleware_internal_test.go.
func (s *Server) authRateLimit() func(http.Handler) http.Handler {
	if s.app.IsTest() {
		return func(next http.Handler) http.Handler { return next }
	}

	return newAuthRateLimit(s.handlers.TooManyRequests)
}

func newAuthRateLimit(limitHandler http.HandlerFunc) func(http.Handler) http.Handler {
	return httprate.LimitBy(
		authAttemptLimit,
		authAttemptWindow,
		keyByClientIP,
		httprate.WithLimitHandler(limitHandler),
	)
}

// keyByClientIP buckets by the connecting address, which realClientIP has
// already replaced with the forwarded client address. httprate's own KeyByIP is
// deprecated precisely because it cannot know whether a proxy sits in front;
// here it does, so the choice is made explicitly.
func keyByClientIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr carries no port when a middleware rewrote it from a header.
		ip = r.RemoteAddr
	}

	return httprate.CanonicalizeIP(ip), nil
}

const xForwardedForHeader = "X-Forwarded-For"

// realClientIP rewrites RemoteAddr from the reverse proxy's X-Forwarded-For so
// the rate limiter buckets on the client rather than on Caddy.
//
// This deliberately does not use chi's middleware.RealIP. That one reads
// True-Client-IP, then X-Real-IP, then the *first* X-Forwarded-For entry, and
// all three are client-controlled behind a stock Caddy: Caddy sets neither of
// the first two, so they arrive verbatim, and it *appends* to X-Forwarded-For
// rather than replacing it, so the first entry is whatever the client sent.
// Trusting any of them lets a client mint a new rate-limit bucket per request
// by rotating a header value. Binding loopback stops direct connections; it
// does nothing about headers arriving through the proxy.
//
// The last X-Forwarded-For entry is the only one the proxy itself wrote, so
// that is the one used here.
func realClientIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ip := forwardedClientIP(r); ip != "" {
			r.RemoteAddr = ip
		}

		next.ServeHTTP(w, r)
	})
}

func forwardedClientIP(r *http.Request) string {
	values := r.Header.Values(xForwardedForHeader)
	if len(values) == 0 {
		return ""
	}

	// Repeated headers are equivalent to a single comma-joined one, so the
	// proxy's own entry is the last element of the last header.
	last := values[len(values)-1]
	if i := strings.LastIndex(last, ","); i >= 0 {
		last = last[i+1:]
	}

	last = strings.TrimSpace(last)
	if net.ParseIP(last) == nil {
		return ""
	}

	return last
}

// setUpMiddlewares registers what every request pays for, static assets
// included. Keep it cheap: anything touching the database belongs in
// setUpAppMiddlewares.
func (s *Server) setUpMiddlewares() {
	// realClientIP runs before the logger so access logs record the client
	// rather than the proxy. It reads only the last X-Forwarded-For entry,
	// which is the one Caddy wrote — see the comment on realClientIP for why
	// the client-supplied entries are not trusted.
	s.Router.Use(realClientIP)

	if !s.app.IsTest() {
		s.Router.Use(middleware.Logger)
	}

	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.RequestID)
	s.Router.Use(s.baseSecurityHeaders)
}

// setUpAppMiddlewares is the chain for routes that render the app. Static assets
// are mounted outside of it, so serving a stylesheet no longer loads the session
// and looks up the current user.
func (s *Server) setUpAppMiddlewares(root chi.Router) {
	root.Use(s.Session.LoadAndSave)
	root.Use(s.limitRequestBody)

	root.Use(s.WithTimeout(5 * time.Second))

	root.Use(s.contentSecurityPolicy)
	root.Use(s.csrf)

	root.Use(s.setTmplData)
	root.Use(s.AuthMiddleware)
}

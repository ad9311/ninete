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
	"github.com/ad9311/ninete/internal/repo"
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
		// Trailing slash trimmed before the lookup: guestRoutes matches exactly,
		// but root.Get("/*") matches "/login/" too, so a bookmark carrying the
		// slash would miss the exemption and bounce a guest away from the
		// login page they asked for.
		path := r.URL.Path
		if path != "/" {
			path = strings.TrimSuffix(path, "/")
		}

		isSignedIn := s.Session.GetBool(r.Context(), handlers.SessionIsUserSignedIn)

		if guestRoutes[path] {
			if isSignedIn {
				http.Redirect(w, r, handlers.AppDashboardPath, http.StatusSeeOther)

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
			http.Redirect(w, r, handlers.AppLoginPath, http.StatusSeeOther)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// newCSRFHandler builds the shared nosurf handler. Both chains use the same
// cookie, so a token minted while rendering a page is the token the API accepts.
func (s *Server) newCSRFHandler(next http.Handler) *nosurf.CSRFHandler {
	csrfHandler := nosurf.New(next)
	// Browsers post CSP reports automatically with no CSRF token.
	csrfHandler.ExemptPath(cspReportPath)
	// Secure is a method call rather than a literal, which gosec cannot follow, so it
	// reads the cookie as missing the attribute. It is set in production and left off
	// locally so the cookie survives plain HTTP in development.
	// #nosec G124 -- HttpOnly and SameSite are set here; Secure follows the environment.
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   s.app.IsProduction(),
		SameSite: http.SameSiteLaxMode,
		Name:     "ninete_csrf",
	})

	return csrfHandler
}

func (s *Server) csrf(next http.Handler) http.Handler {
	return s.newCSRFHandler(next)
}

// apiCSRF is the same protection with a JSON rejection, so a client parsing the
// body sees the standard error envelope instead of nosurf's plain text.
//
// A bearer-authenticated request is exempt. CSRF exists because a browser
// attaches the session cookie to a request another site triggers; nothing
// attaches an Authorization header on its own, so a request carrying a valid
// token cannot have been forged that way. The exemption keys off the token
// tokenAuth put in the context — never off the header's mere presence — so a
// request with a bad token was already answered 401 and never gets here.
func (s *Server) apiCSRF(next http.Handler) http.Handler {
	csrfHandler := s.newCSRFHandler(next)
	csrfHandler.SetFailureHandler(http.HandlerFunc(s.handlers.APIForbidden))
	csrfHandler.ExemptFunc(isTokenAuthenticated)

	return csrfHandler
}

func isTokenAuthenticated(r *http.Request) bool {
	token, ok := r.Context().Value(handlers.KeyAPIToken).(*repo.APIToken)

	return ok && token != nil
}

// tokenAPIPrefixes are the only /api paths a bearer token may reach, matched
// as whole path segments. An allowlist rather than a denylist on purpose: a
// route added later stays browser-only until someone decides a token should
// have it, instead of being open to tokens by default. /api/tokens,
// /api/delete-data, /api/login and /api/register are left out deliberately —
// a token cannot manage tokens, wipe data or open a session.
var tokenAPIPrefixes = []string{ //nolint:gochecknoglobals // static allowlist
	apiPathPrefix + "/session",
	apiPathPrefix + "/categories",
	apiPathPrefix + "/dashboard",
	apiPathPrefix + "/report-settings",
	apiPathPrefix + "/tags",
	apiPathPrefix + "/expenses",
	apiPathPrefix + "/recurrent-expenses",
}

func tokenMayReach(path string) bool {
	for _, prefix := range tokenAPIPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}

// tokenMethodAllowed maps a scope to the methods it grants. DELETE is in
// neither: no token can remove anything, whatever its scope.
func tokenMethodAllowed(scope, method string) bool {
	switch method {
	case http.MethodGet:
		return scope == logic.APITokenScopeRead || scope == logic.APITokenScopeWrite
	case http.MethodPost, http.MethodPut:
		return scope == logic.APITokenScopeWrite
	default:
		return false
	}
}

const bearerScheme = "bearer "

// tokenAuth authenticates "Authorization: Bearer <token>" on the API chain. It
// runs before apiCSRF and apiAuth, and a request without the header passes
// through untouched to the session path.
//
// Once the header is present the session is out of the picture for good: a
// malformed, unknown, expired or revoked token answers 401 and never falls
// back to the cookie. Falling back would let a request with a junk header and
// an ambient session cookie skip CSRF — the exemption would be keyed off the
// wrong credential.
//
// failures throttles bad tokens per client; nil disables it, which is what
// ENV=test uses. Only failures count, so a working client is never limited.
func (s *Server) tokenAuth(failures *tokenFailureLimit) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)

				return
			}

			key, _ := keyByClientIP(r)
			if failures.blocked(key) {
				s.handlers.APITooManyRequests(w, r)

				return
			}

			unauthorized := func() {
				failures.record(w, r, key)
				w.Header().Set("WWW-Authenticate", `Bearer realm="ninete"`)
				s.handlers.APIUnauthorized(w, r)
			}

			if len(header) <= len(bearerScheme) || !strings.EqualFold(header[:len(bearerScheme)], bearerScheme) {
				unauthorized()

				return
			}

			ctx := r.Context()

			token, err := s.store.AuthenticateAPIToken(ctx, strings.TrimSpace(header[len(bearerScheme):]))
			if err != nil {
				if errors.Is(err, logic.ErrAPITokenInvalid) {
					unauthorized()

					return
				}

				s.app.Logger.Errorf("failed to authenticate api token %v", err)
				s.handlers.APIInternalError(w, r)

				return
			}

			if !tokenMayReach(r.URL.Path) {
				s.handlers.WriteJSONError(w, http.StatusForbidden, handlers.ErrTokenNotAllowed)

				return
			}

			if !tokenMethodAllowed(token.Scope, r.Method) {
				if r.Method == http.MethodPost || r.Method == http.MethodPut {
					s.handlers.WriteJSONError(w, http.StatusForbidden, handlers.ErrTokenScope)

					return
				}

				s.handlers.WriteJSONError(w, http.StatusForbidden, handlers.ErrTokenNotAllowed)

				return
			}

			user, err := s.store.FindUser(ctx, token.UserID)
			if err != nil {
				// The foreign key cascades, so a token whose user is gone
				// cannot normally exist; treat one as a dead credential.
				if errors.Is(err, sql.ErrNoRows) {
					unauthorized()

					return
				}

				s.app.Logger.Errorf("failed to find api token user %v", err)
				s.handlers.APIInternalError(w, r)

				return
			}

			ctx = context.WithValue(ctx, handlers.KeyAPIToken, &token)
			ctx = context.WithValue(ctx, handlers.KeyCurrentUser, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Failed bearer attempts allowed per client per window. A configured client
// never fails; something cycling through tokens hits this on the first burst.
const (
	tokenFailureLimitCount  = 10
	tokenFailureLimitWindow = time.Minute
)

// tokenFailureLimit counts failed bearer attempts per client. It is separate
// from authRateLimit on purpose: that value is shared by the two credential
// routes and must not gain a third caller (the CLAUDE.md invariant), and it
// counts every attempt, where this counts only failures.
type tokenFailureLimit struct {
	limiter *httprate.RateLimiter
}

func newTokenFailureLimit() *tokenFailureLimit {
	return &tokenFailureLimit{
		limiter: httprate.NewRateLimiter(tokenFailureLimitCount, tokenFailureLimitWindow),
	}
}

// tokenFailures is the server's limiter, or nil under ENV=test, where the
// failure paths are exercised far more often than a real client would.
// TestTokenFailureLimit covers the limiter directly.
func (s *Server) tokenFailures() *tokenFailureLimit {
	if s.app.IsTest() {
		return nil
	}

	return newTokenFailureLimit()
}

func (l *tokenFailureLimit) blocked(key string) bool {
	if l == nil {
		return false
	}

	_, rate, err := l.limiter.Status(key)
	if err != nil {
		return true
	}

	return rate >= tokenFailureLimitCount
}

func (l *tokenFailureLimit) record(w http.ResponseWriter, r *http.Request, key string) {
	if l == nil {
		return
	}

	l.limiter.OnLimit(w, r, key)
}

// apiAuth is the API's answer to AuthMiddleware: 401 with a JSON body instead
// of a redirect (§3.1). It also puts the signed-in user into KeyCurrentUser,
// which the API chain would otherwise never set — that happens in setTmplData,
// which the API chain deliberately skips, and every resource handler opens with
// getCurrentUser, which panics when the key is absent.
func (s *Server) apiAuth(next http.Handler) http.Handler {
	// PostAPILogin/PostAPIRegister (Phase 6 of docs/spa-migration.md). Built
	// once per chain, like AuthMiddleware's own guest set.
	guestAPIRoutes := map[string]bool{
		"/api/login":    true,
		"/api/register": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// tokenAuth already authenticated this request and set the user; the
		// session must not be consulted for it at all.
		if guestAPIRoutes[r.URL.Path] || isTokenAuthenticated(r) {
			next.ServeHTTP(w, r)

			return
		}

		ctx := r.Context()
		if !s.Session.GetBool(ctx, handlers.SessionIsUserSignedIn) {
			s.handlers.APIUnauthorized(w, r)

			return
		}

		user, err := s.store.FindUser(ctx, s.Session.GetInt(ctx, handlers.SessionUserID))
		if err != nil {
			// A session pointing at a user who no longer exists is a stale
			// credential, not a server fault.
			if errors.Is(err, sql.ErrNoRows) {
				s.handlers.APIUnauthorized(w, r)

				return
			}

			s.app.Logger.Errorf("failed to find current user %v", err)
			s.handlers.APIInternalError(w, r)

			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, handlers.KeyCurrentUser, &user)))
	})
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
			"appBundle":      s.assetPath("app"),
			"appStylesheet":  s.assetPath("css"),
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

// apiPathPrefix is the JSON chain's mount point. Matched as a whole path
// segment wherever it is used, so a future "/api-tokens" page route is not
// mistaken for an API route.
const apiPathPrefix = "/api"

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
// The returned middleware is shared by both routes it guards — /api/login and
// /api/register — so a client draws on one budget rather than one per route.
// Call it once and pass the value around; calling it again builds an
// independent counter and multiplies the allowance.
//
// Disabled under ENV=test: the suite performs a hundred logins from one
// synthetic address, which is exactly the pattern this blocks. The middleware
// itself is covered directly in middleware_internal_test.go.
func (s *Server) authRateLimit() func(http.Handler) http.Handler {
	if s.app.IsTest() {
		return func(next http.Handler) http.Handler { return next }
	}

	return newAuthRateLimit(s.handlers.APITooManyRequests)
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

// requestTimeout bounds a request in either chain. Exports materialize their
// whole payload inside it — see docs/spa-migration.md §3.3.
const requestTimeout = 5 * time.Second

// setUpAppMiddlewares is the chain for routes that render the app. Static assets
// are mounted outside of it, so serving a stylesheet no longer loads the session
// and looks up the current user.
func (s *Server) setUpAppMiddlewares(root chi.Router) {
	root.Use(s.Session.LoadAndSave)
	root.Use(s.limitRequestBody)

	root.Use(s.WithTimeout(requestTimeout))

	root.Use(s.contentSecurityPolicy)
	root.Use(s.csrf)

	root.Use(s.setTmplData)
	root.Use(s.AuthMiddleware)
}

// setUpAPIMiddlewares is the chain for the JSON API. It shares the session, the
// body cap, the timeout and the CSRF protection of the page chain, and drops
// the two pieces that assume HTML: setTmplData (no templates, no template map)
// and AuthMiddleware (which redirects rather than answering 401).
//
// There is no CSP here on purpose: the policy exists to constrain a rendered
// document, and a JSON response has none. baseSecurityHeaders still applies, so
// the responses keep nosniff.
func (s *Server) setUpAPIMiddlewares(api chi.Router) {
	api.Use(s.Session.LoadAndSave)
	api.Use(s.limitRequestBody)
	api.Use(s.WithTimeout(requestTimeout))
	// tokenAuth must run before apiCSRF: the CSRF exemption reads the token it
	// puts in the context.
	api.Use(s.tokenAuth(s.tokenFailures()))
	api.Use(s.apiCSRF)
	api.Use(s.apiAuth)
}

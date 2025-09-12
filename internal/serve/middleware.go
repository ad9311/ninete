package serve

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/go-chi/chi/v5/middleware"
)

// WithTimeout sets a timeout for each request using the provided duration.
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

// JSONMiddleware enforces that requests with a body must send Content-Type: application/json
// and ensures all responses declare Content-Type: application/json.
func (s *Server) JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			ct := r.Header.Get("Content-Type")
			if ct == "" || !strings.HasPrefix(ct, "application/json") {
				s.respondError(
					w,
					http.StatusUnsupportedMediaType,
					CodeGeneric,
					ErrContentNotSupported,
				)

				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// CORS sets the CORS headers and handles preflight requests.
func (s *Server) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		isOrigin := slices.Contains(s.allowedOrigins, origin)

		isPreflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""

		if isOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
		}

		if isPreflight {
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")

			if !isOrigin {
				s.respondError(w, http.StatusForbidden, CodeForbidden, ErrOriginNotAllowed)

				return
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

			reqHeaders := r.Header.Get("Access-Control-Request-Headers")
			if reqHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
			} else {
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}

			w.Header().Set("Access-Control-Max-Age", "1800")

			w.WriteHeader(http.StatusNoContent)

			return
		}

		if r.Method == http.MethodOptions {
			if !isOrigin {
				s.respondError(w, http.StatusForbidden, CodeForbidden, ErrOriginNotAllowed)

				return
			}
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// NotFoundHandler returns a JSON error response for unknown routes.
func (s *Server) NotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	// writeError(w, http.StatusNotFound, routeNotFound, errs.ErrNotFound)
	s.respondError(w, http.StatusNotFound, CodeGeneric, ErrNotPathFound)
}

// MethodNotAllowedHandler returns a JSON error response when the HTTP method is not allowed.
func (s *Server) MethodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	s.respondError(w, http.StatusMethodNotAllowed, CodeForbidden, ErrMethodNotAllowed)
}

func (s *Server) setUpMiddlewares() {
	if s.app.ENV != prog.ENVTest {
		s.Router.Use(middleware.Logger)
	}

	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.RequestID)

	s.Router.Use(s.WithTimeout(5 * time.Second))
	s.Router.Use(s.JSONMiddleware)
	s.Router.Use(s.CORS)

	s.Router.NotFound(s.NotFoundHandler)
	s.Router.MethodNotAllowed(s.MethodNotAllowedHandler)
}

package server

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/errs"
)

// WithTimeout sets a timeout for each request using the provided duration.
func WithTimeout(dur time.Duration) func(http.Handler) http.Handler {
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
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			ct := r.Header.Get("Content-Type")
			if ct == "" || !strings.HasPrefix(ct, "application/json") {
				writeError(w, http.StatusUnsupportedMediaType, standardErrorCode, errs.ErrUnsupportedMediaType)

				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// NotFoundHandler returns a JSON error response for unknown routes.
func NotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, routeNotFound, errs.ErrNotFound)
}

// MethodNotAllowedHandler returns a JSON error response when the HTTP method is not allowed.
func MethodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusMethodNotAllowed, methodNotAllowed, errs.ErrMethodNotAllowed)
}

// CORS sets the CORS headers and handles preflight requests.
func (s *Server) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		isOrigin := slices.Contains(s.config.AllowedOrigins, origin)

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
				writeError(w, http.StatusForbidden, standardErrorCode, errs.ErrMethodNotAllowedForOrigin)

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
				writeError(w, http.StatusForbidden, standardErrorCode, errs.ErrMethodNotAllowedForOrigin)

				return
			}
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware verifies and validates the access token for protected routes. Adds user info to context if valid.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, invalidAuthCredsErrorCode, errs.ErrInvalidAuthHeader)

			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := s.serviceStore.ParseAndValidateJWT(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, invalidAuthCredsErrorCode, err)

			return
		}

		userID, ok := claims["sub"].(float64)
		if !ok {
			writeError(w, http.StatusUnauthorized, standardErrorCode, errs.ErrInvalidClaimsType)

			return
		}

		user, err := s.serviceStore.FindUserByID(r.Context(), int32(userID))
		if err != nil {
			writeError(w, http.StatusUnauthorized, standardErrorCode, err)

			return
		}

		ctx := context.WithValue(r.Context(), app.CurrentUserIDKey, int32(userID))
		ctx = context.WithValue(ctx, app.CurrentUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

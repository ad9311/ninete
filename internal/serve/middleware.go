package serve

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/go-chi/chi/v5/middleware"
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

func (s *Server) NotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	s.respondError(w, http.StatusNotFound, CodeGeneric, ErrNotPathFound)
}

func (s *Server) MethodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	s.respondError(w, http.StatusMethodNotAllowed, CodeForbidden, ErrMethodNotAllowed)
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			s.respondError(w, http.StatusUnauthorized, CodeGeneric, ErrInvalidAccessToken)

			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := s.store.ParseAndValidateJWT(tokenString)
		if err != nil {
			s.respondError(w, http.StatusUnauthorized, CodeGeneric, err)

			return
		}

		userIDStr, ok := claims["sub"].(string)
		if !ok {
			err := fmt.Errorf("%w, invalid claims sub type", logic.ErrInvalidJWTToken)
			s.respondError(w, http.StatusUnauthorized, CodeBadFormat, err)

			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			err := fmt.Errorf("%w, invalid claims sub value", logic.ErrInvalidJWTToken)
			s.respondError(w, http.StatusUnauthorized, CodeBadFormat, err)

			return
		}

		user, err := s.store.FindUserByID(r.Context(), userID)
		if err != nil {
			s.respondError(w, http.StatusUnauthorized, CodeGeneric, err)

			return
		}

		ctx := context.WithValue(r.Context(), prog.KeyCurrentUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

package serve

import (
	"context"
	"database/sql"
	"errors"
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isUserSignedIn := s.Session.GetBool(r.Context(), handlers.SessionIsUserSignedIn)
		requestPath := r.URL.Path

		if isUserSignedIn {
			if requestPath == "/login" || requestPath == "/register" {
				http.Redirect(w, r, "/", http.StatusSeeOther)

				return
			}

			next.ServeHTTP(w, r)

			return
		}

		for _, route := range publicRoutes() {
			if strings.HasPrefix(requestPath, route) {
				next.ServeHTTP(w, r)

				return
			}
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}

func (s *Server) csrf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
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

		templateMap := map[string]any{
			"csrfToken":      csrf,
			"error":          "",
			"isUserSignedIn": isUserSignedIn,
			"currentUser":    currentUser,
		}

		ctx = context.WithValue(ctx, handlers.KeyCurrentUser, currentUser)
		ctx = context.WithValue(ctx, handlers.KeyTemplateData, templateMap)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) setUpMiddlewares() {
	if !s.app.IsTest() {
		s.Router.Use(middleware.Logger)
	}

	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.RequestID)

	s.Router.Use(s.WithTimeout(5 * time.Second))

	s.Router.Use(s.csrf)

	s.Router.Use(s.AuthMiddleware)

	s.Router.Use(s.setTmplData)

	s.Router.NotFound(s.handlers.NotFound)
	s.Router.MethodNotAllowed(s.handlers.MethodNotAllowed)
}

func publicRoutes() []string {
	return []string{
		"/login",
		"/register",
		"/static",
	}
}

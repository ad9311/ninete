package serve

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/webkeys"
	"github.com/ad9311/ninete/internal/webtmpl"
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

func (s *Server) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, http.StatusNotFound, webtmpl.NotFoundIndex, s.tmplData(r))
}

func (s *Server) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	data := s.tmplData(r)
	data["error"] = ErrNotAllowed.Error()
	s.renderTemplate(w, http.StatusMethodNotAllowed, webtmpl.ErrorIndex, data)
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isUserSignedIn := s.Session.GetBool(r.Context(), webkeys.SessionIsUserSignedIn)
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

		var currentUser repo.User
		id := s.Session.GetInt(ctx, webkeys.SessionUserID)
		if id > 0 {
			currentUser, _ = s.store.FindUser(ctx, id)
		}

		templateMap := map[string]any{
			"csrfToken":   csrf,
			"error":       "",
			"currentUser": currentUser,
		}

		ctx = context.WithValue(ctx, webkeys.TemplateData, templateMap)
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

	s.Router.NotFound(s.NotFoundHandler)
	s.Router.MethodNotAllowed(s.MethodNotAllowedHandler)
}

func publicRoutes() []string {
	return []string{
		"/login",
		"/register",
		"/static",
	}
}

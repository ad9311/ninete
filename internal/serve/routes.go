package serve

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) setUpRoutes() {
	s.Router.Route("/", func(root chi.Router) {
		root.Get("/", s.handlers.GetRoot)

		root.Group(func(auth chi.Router) {
			auth.Get("/login", s.handlers.GetLogin)
			auth.Post("/login", s.handlers.PostLogin)
			auth.Post("/logout", s.handlers.PostLogout)
		})

		root.Group(func(dashboard chi.Router) {
			dashboard.Get("/dashboard", s.handlers.GetDashboard)
		})

		setUpFileServer(root)
	})
}

func setUpFileServer(root chi.Router) {
	fileServer := http.FileServer(http.Dir("./web/static/"))
	root.Handle("/static/*", http.StripPrefix("/static/", fileServer))
}

func (s *Server) setUpSession() {
	s.Session.Cookie.Secure = s.app.IsProduction()
	s.Session.Cookie.Name = "ninete_session"
}

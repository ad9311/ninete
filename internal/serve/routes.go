package serve

import "github.com/go-chi/chi/v5"

func (s *Server) setUpRoutes() {
	s.Router.Route("/", func(root chi.Router) {
		root.Group(func(status chi.Router) {
			status.Get("/healthz", s.GetHealthz)
			status.Get("/readyz", s.GetReadyz)
		})

		root.Group(func(auth chi.Router) {
			auth.Route("/auth", func(auth chi.Router) {
				auth.Post("/sign-up", s.PostSignUp)
				auth.Post("/sign-in", s.PostSignIn)
				auth.Delete("/sign-out", s.DeleteSignOut)
				auth.Post("/refresh", s.PostRefresh)
			})
		})

		root.Group(func(secure chi.Router) {
			secure.Use(s.AuthMiddleware)

			secure.Get("/users/me", s.GetMe)
		})
	})
}

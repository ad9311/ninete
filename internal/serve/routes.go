package serve

import "github.com/go-chi/chi/v5"

func (s *Server) setUpRoutes() {
	s.Router.Route("/", func(root chi.Router) {
		root.Group(func(status chi.Router) {
			status.Get("/healthz", s.getHealthz)
			status.Get("/readyz", s.getReadyz)
		})

		root.Group(func(auth chi.Router) {
			auth.Route("/auth", func(auth chi.Router) {
				auth.Post("/sign-up", s.postSignUp)
				auth.Post("/sign-in", s.postSignIn)
				auth.Delete("/sign-out", s.deleteSignOut)
			})
		})

		root.Group(func(secure chi.Router) {
			secure.Use(s.AuthMiddleware)

			secure.Get("/users/me", s.getMe)
		})
	})
}

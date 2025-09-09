package server

import (
	"github.com/go-chi/chi/v5"
)

// setupRoutes configures all HTTP routes for the server, including health checks,
// authentication, and protected user endpoints.
func (s *Server) setupRoutes() {
	s.Router.Route("/", func(root chi.Router) {
		// Health, readiness and metrics
		root.Get("/healthz", s.GetHealthz)
		root.Get("/readyz", s.GetReadyz)

		// Authentication
		root.Route("/auth", func(auth chi.Router) {
			auth.Post("/sign-up", s.PostSignUp)
			auth.Post("/sign-in", s.PostSignIn)
			auth.Post("/refresh", s.PostRefresh)
		})

		root.Group(func(secure chi.Router) {
			secure.Use(s.AuthMiddleware)

			// Protected auth
			secure.Delete("/auth/sign-out", s.DeleteSignOut)

			// Users
			secure.Get("/users/me", s.GetMe)
		})
	})
}

package serve

import "github.com/go-chi/chi/v5"

func (s *Server) setUpRoutes() {
	s.Router.Route("/", func(root chi.Router) {
		root.Get("/healthz", s.getHealthz)
		root.Get("/readyz", s.getReadyz)
	})
}

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

			secure.Get("/categories", s.GetCategories)

			secure.Route("/expenses", func(e chi.Router) {
				e.Get("/", s.GetExpenses)
				e.Post("/", s.PostExpense)
				e.Route("/{id}", func(e chi.Router) {
					e.Use(s.ContextExpense)
					e.Get("/", s.GetExpense)
					e.Put("/", s.PutExpense)
					e.Patch("/", s.PutExpense)
					e.Delete("/", s.DeleteExpense)
				})
			})

			secure.Route("/recurrent-expenses", func(re chi.Router) {
				re.Get("/", s.GetRecurrentExpenses)
				re.Post("/", s.PostRecurrentExpense)
				re.Route("/{id}", func(e chi.Router) {
					e.Use(s.ContextRecurrentExpense)
					e.Get("/", s.GetRecurrentExpense)
					e.Put("/", s.PutRecurrentExpense)
					e.Patch("/", s.PutRecurrentExpense)
					e.Delete("/", s.DeleteRecurrentExpense)
				})
			})
		})
	})
}

package serve

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) setUpRoutes() {
	s.Router.Route("/", func(root chi.Router) {
		root.Get("/", s.handlers.GetRoot)

		root.Group(func(auth chi.Router) {
			auth.Get("/register", s.handlers.GetRegister)
			auth.Post("/register", s.handlers.PostRegister)
			auth.Get("/login", s.handlers.GetLogin)
			auth.Post("/login", s.handlers.PostLogin)
			auth.Post("/logout", s.handlers.PostLogout)
		})

		root.Group(func(dashboard chi.Router) {
			dashboard.Get("/dashboard", s.handlers.GetDashboard)
		})

		root.Group(func(expenses chi.Router) {
			expenses.Route("/expenses", func(expenses chi.Router) {
				expenses.Get("/", s.handlers.GetExpenses)
				expenses.Post("/", s.handlers.PostExpenses)
				expenses.Get("/new", s.handlers.GetExpensesNew)
				expenses.Route("/{id}", func(expenses chi.Router) {
					expenses.Use(s.handlers.ExpenseContext)

					expenses.Get("/", s.handlers.GetExpense)
					expenses.Post("/", s.handlers.PostExpensesUpdate)
					expenses.Get("/edit", s.handlers.GetExpensesEdit)
					expenses.Post("/delete", s.handlers.PostExpensesDelete)
				})
			})
		})

		root.Group(func(recurrentExpenses chi.Router) {
			recurrentExpenses.Route("/recurrent-expenses", func(recurrentExpenses chi.Router) {
				recurrentExpenses.Get("/", s.handlers.GetRecurrentExpenses)
				recurrentExpenses.Post("/", s.handlers.PostRecurrentExpenses)
				recurrentExpenses.Get("/new", s.handlers.GetRecurrentExpensesNew)
				recurrentExpenses.Route("/{id}", func(recurrentExpenses chi.Router) {
					recurrentExpenses.Use(s.handlers.RecurrentExpenseContext)

					recurrentExpenses.Post("/", s.handlers.PostRecurrentExpensesUpdate)
					recurrentExpenses.Get("/edit", s.handlers.GetRecurrentExpensesEdit)
					recurrentExpenses.Post("/delete", s.handlers.PostRecurrentExpensesDelete)
				})
			})
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

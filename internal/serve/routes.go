package serve

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) setUpRoutes() {
	s.Router.Route("/", func(root chi.Router) {
		setUpFileServer(root)

		root.Get("/", s.handlers.GetRoot)

		root.Get("/login", s.handlers.GetLogin)
		root.Post("/login", s.handlers.PostLogin)
		root.Get("/register", s.handlers.GetRegister)
		root.Post("/register", s.handlers.PostRegister)
		root.Post("/logout", s.handlers.PostLogout)

		root.Get("/dashboard", s.handlers.GetDashboard)

		root.Route("/expenses", func(expenses chi.Router) {
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

		root.Route("/recurrent-expenses", func(recurrentExpenses chi.Router) {
			recurrentExpenses.Get("/", s.handlers.GetRecurrentExpenses)
			recurrentExpenses.Post("/", s.handlers.PostRecurrentExpenses)
			recurrentExpenses.Get("/new", s.handlers.GetRecurrentExpensesNew)
			recurrentExpenses.Route("/{id}", func(recurrentExpenses chi.Router) {
				recurrentExpenses.Use(s.handlers.RecurrentExpenseContext)

				recurrentExpenses.Get("/", s.handlers.GetRecurrentExpense)
				recurrentExpenses.Post("/", s.handlers.PostRecurrentExpensesUpdate)
				recurrentExpenses.Get("/edit", s.handlers.GetRecurrentExpensesEdit)
				recurrentExpenses.Post("/delete", s.handlers.PostRecurrentExpensesDelete)
			})
		})

		root.Route("/lists", func(lists chi.Router) {
			lists.Get("/", s.handlers.GetLists)
			lists.Post("/", s.handlers.PostLists)
			lists.Get("/new", s.handlers.GetListsNew)
			lists.Route("/{id}", func(lists chi.Router) {
				lists.Use(s.handlers.ListContext)

				lists.Get("/", s.handlers.GetList)
				lists.Post("/", s.handlers.PostListsUpdate)
				lists.Get("/edit", s.handlers.GetListsEdit)
				lists.Post("/delete", s.handlers.PostListsDelete)

				lists.Route("/tasks", func(tasks chi.Router) {
					tasks.Post("/", s.handlers.PostTasks)
					tasks.Get("/new", s.handlers.GetTasksNew)
					tasks.Route("/{task_id}", func(tasks chi.Router) {
						tasks.Use(s.handlers.TaskContext)

						tasks.Get("/", s.handlers.GetTask)
						tasks.Post("/", s.handlers.PostTasksUpdate)
						tasks.Get("/edit", s.handlers.GetTasksEdit)
						tasks.Post("/delete", s.handlers.PostTasksDelete)
						tasks.Post("/done", s.handlers.PostTasksDone)
					})
				})
			})
		})
	})
}

func setUpFileServer(root chi.Router) {
	fileServer := http.FileServer(http.Dir("./web/static/"))
	root.Handle("/static/*", http.StripPrefix("/static/", fileServer))
}

func (s *Server) setUpSession() {
	s.Session.Lifetime = 7 * 24 * time.Hour
	s.Session.IdleTimeout = 2 * time.Hour
	s.Session.Cookie.Secure = s.app.IsProduction()
	s.Session.Cookie.HttpOnly = true
	s.Session.Cookie.Persist = true
	s.Session.Cookie.SameSite = http.SameSiteStrictMode
	s.Session.Cookie.Name = "ninete_session"
}

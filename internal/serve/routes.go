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

		root.Route("/exports", func(exports chi.Router) {
			exports.Get("/", s.handlers.GetExports)
			exports.Get("/expenses.json", s.handlers.GetExportsExpenses)
		})

		root.Route("/expenses", func(expenses chi.Router) {
			expenses.Get("/", s.handlers.GetExpenses)
			expenses.Post("/", s.handlers.PostExpenses)
			expenses.Get("/new", s.handlers.GetExpensesNew)
			expenses.Get("/stats", s.handlers.GetExpensesStats)
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

		root.Route("/macros", func(r chi.Router) {
			r.Get("/", s.handlers.GetMacros)
			r.Post("/", s.handlers.PostMacros)
			r.Get("/new", s.handlers.GetMacrosNew)
			r.Get("/goals", s.handlers.GetMacrosGoals)
			r.Post("/goals", s.handlers.PostMacrosGoals)
			r.Get("/stats", s.handlers.GetMacrosStats)
			r.Route("/templates", func(r chi.Router) {
				r.Get("/", s.handlers.GetMacroTemplates)
				r.Post("/", s.handlers.PostMacroTemplates)
				r.Get("/new", s.handlers.GetMacroTemplatesNew)
				r.Route("/{template_id}", func(r chi.Router) {
					r.Use(s.handlers.MacroTemplateContext)
					r.Get("/", s.handlers.GetMacroTemplate)
					r.Post("/", s.handlers.PostMacroTemplateUpdate)
					r.Get("/edit", s.handlers.GetMacroTemplateEdit)
					r.Post("/delete", s.handlers.PostMacroTemplateDelete)
				})
			})
			r.Route("/{id}", func(r chi.Router) {
				r.Use(s.handlers.MacroEntryContext)
				r.Get("/", s.handlers.GetMacroEntry)
				r.Post("/", s.handlers.PostMacroEntryUpdate)
				r.Get("/edit", s.handlers.GetMacroEntryEdit)
				r.Post("/delete", s.handlers.PostMacroEntryDelete)
			})
		})

		root.Route("/foods", func(foods chi.Router) {
			foods.Get("/", s.handlers.GetFoods)
			foods.Post("/", s.handlers.PostFoods)
			foods.Get("/new", s.handlers.GetFoodsNew)
			foods.Route("/{food_id}", func(foods chi.Router) {
				foods.Use(s.handlers.FoodContext)

				foods.Get("/", s.handlers.GetFood)
				foods.Post("/", s.handlers.PostFoodUpdate)
				foods.Get("/edit", s.handlers.GetFoodEdit)
				foods.Post("/delete", s.handlers.PostFoodDelete)
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
	s.Session.Cookie.Secure = s.app.IsProduction()
	s.Session.Cookie.HttpOnly = true
	s.Session.Cookie.Persist = true
	s.Session.Cookie.SameSite = http.SameSiteStrictMode
	s.Session.Cookie.Name = "ninete_session"
}

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

		root.Post(cspReportPath, s.handlers.PostCSPReport)

		root.Get("/login", s.handlers.GetLogin)
		root.Post("/login", s.handlers.PostLogin)
		root.Get("/register", s.handlers.GetRegister)
		root.Post("/register", s.handlers.PostRegister)
		root.Post("/logout", s.handlers.PostLogout)

		root.Get("/dashboard", s.handlers.GetDashboard)

		root.Route("/account", func(account chi.Router) {
			account.Get("/", s.handlers.GetAccount)
			account.Post("/expenses/delete-all", s.handlers.PostAccountDeleteExpenses)
			account.Post("/recurrent-expenses/delete-all", s.handlers.PostAccountDeleteRecurrentExpenses)
			account.Post("/macro-entries/delete-all", s.handlers.PostAccountDeleteMacroEntries)
			account.Post("/macro-goals/delete-all", s.handlers.PostAccountDeleteMacroGoals)
			account.Post("/foods/delete-all", s.handlers.PostAccountDeleteFoods)
			account.Post("/moods/delete-all", s.handlers.PostAccountDeleteMoodEntries)
			account.Post("/tags/delete-all", s.handlers.PostAccountDeleteTags)
			account.Post("/delete-all", s.handlers.PostAccountDeleteAll)
		})

		root.Route("/exports", func(exports chi.Router) {
			exports.Get("/", s.handlers.GetExports)
			exports.Get("/expenses.json", s.handlers.GetExportsExpenses)
		})

		root.Route("/expenses", func(expenses chi.Router) {
			expenses.Get("/", s.handlers.GetExpenses)
			expenses.Post("/", s.handlers.PostExpenses)
			expenses.Post("/quick", s.handlers.PostExpensesQuick)
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
			foods.Route("/{id}", func(foods chi.Router) {
				foods.Use(s.handlers.FoodContext)

				foods.Get("/", s.handlers.GetFood)
				foods.Post("/", s.handlers.PostFoodUpdate)
				foods.Get("/edit", s.handlers.GetFoodEdit)
				foods.Post("/delete", s.handlers.PostFoodDelete)
			})
		})

		root.Route("/moods", func(moods chi.Router) {
			moods.Get("/", s.handlers.GetMoodEntries)
			moods.Post("/", s.handlers.PostMoodEntries)
			moods.Get("/new", s.handlers.GetMoodEntriesNew)
			moods.Get("/stats", s.handlers.GetMoodEntriesStats)
			moods.Route("/{id}", func(moods chi.Router) {
				moods.Use(s.handlers.MoodEntryContext)

				moods.Get("/", s.handlers.GetMoodEntry)
				moods.Post("/", s.handlers.PostMoodEntriesUpdate)
				moods.Get("/edit", s.handlers.GetMoodEntriesEdit)
				moods.Post("/delete", s.handlers.PostMoodEntriesDelete)
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
	s.Session.Cookie.SameSite = http.SameSiteLaxMode
	s.Session.Cookie.Name = "ninete_session"
}

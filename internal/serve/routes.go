package serve

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) setUpRoutes() {
	s.setUpFileServer()
	s.setUpAPIRoutes()

	s.Router.Group(func(root chi.Router) {
		s.setUpAppMiddlewares(root)

		// Registered on the group so the fallbacks still get template data.
		root.NotFound(s.handlers.NotFound)
		root.MethodNotAllowed(s.handlers.MethodNotAllowed)

		root.Get("/", s.handlers.GetRoot)

		root.Post(cspReportPath, s.handlers.PostCSPReport)

		root.Get("/login", s.handlers.GetLogin)
		root.Get("/register", s.handlers.GetRegister)
		root.Post("/logout", s.handlers.PostLogout)

		// Only the routes that check a credential are throttled. Rendering the
		// forms stays free. Both routes share one middleware value, so a client
		// gets a single budget across them instead of one each.
		credentialLimit := s.authRateLimit()
		root.With(credentialLimit).Post("/login", s.handlers.PostLogin)
		root.With(credentialLimit).Post("/register", s.handlers.PostRegister)

		root.Get("/dashboard", s.handlers.GetDashboard)

		// The SPA shell (docs/spa-migration.md, Phase 1). Wildcarded so every
		// nested client route resolves on a hard refresh; AuthMiddleware guards
		// it the same as any other page since "/app*" is not a guest route.
		root.Get("/app", s.handlers.GetApp)
		root.Get("/app/*", s.handlers.GetApp)

		root.Route("/account", func(account chi.Router) {
			account.Get("/", s.handlers.GetAccount)
			account.Get("/delete-data", s.handlers.GetDeleteData)

			account.Post("/expenses/delete-all", s.handlers.PostDeleteDataExpenses)
			account.Post("/recurrent-expenses/delete-all", s.handlers.PostDeleteDataRecurrentExpenses)
			account.Post("/expense-budgets/delete-all", s.handlers.PostDeleteDataExpenseBudgets)
			account.Post("/tags/delete-all", s.handlers.PostDeleteDataTags)
			account.Post("/delete-all", s.handlers.PostDeleteDataAll)

			account.Route("/exports", func(exports chi.Router) {
				exports.Get("/", s.handlers.GetExports)
				exports.Get("/expenses.json", s.handlers.GetExportsExpenses)
			})
		})

		root.Route("/expenses", func(expenses chi.Router) {
			expenses.Get("/", s.handlers.GetExpenses)
			expenses.Post("/", s.handlers.PostExpenses)
			expenses.Post("/quick", s.handlers.PostExpensesQuick)
			expenses.Get("/new", s.handlers.GetExpensesNew)
			expenses.Get("/stats", s.handlers.GetExpensesStats)
			expenses.Get("/budgets", s.handlers.GetExpensesBudgets)
			expenses.Post("/budgets", s.handlers.PostExpensesBudgets)
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
			recurrentExpenses.Get("/archived", s.handlers.GetRecurrentExpensesArchived)
			recurrentExpenses.Route("/{id}", func(recurrentExpenses chi.Router) {
				recurrentExpenses.Use(s.handlers.RecurrentExpenseContext)

				recurrentExpenses.Get("/", s.handlers.GetRecurrentExpense)
				recurrentExpenses.Post("/", s.handlers.PostRecurrentExpensesUpdate)
				recurrentExpenses.Get("/edit", s.handlers.GetRecurrentExpensesEdit)
				recurrentExpenses.Post("/delete", s.handlers.PostRecurrentExpensesDelete)
				recurrentExpenses.Post("/unarchive", s.handlers.PostRecurrentExpensesUnarchive)
			})
		})
	})
}

// setUpAPIRoutes mounts the JSON API the SPA talks to. It is a sibling of the
// page group, not a child: the two chains differ (see setUpAPIMiddlewares), and
// an /api route must never fall through to a rendered template.
//
// Resource routes arrive from Phase 2 of docs/spa-migration.md onward; the one
// route here is what the Phase 1 shell reads its current user from.
func (s *Server) setUpAPIRoutes() {
	s.Router.Route("/api", func(api chi.Router) {
		s.setUpAPIMiddlewares(api)

		// Registered on the group so an unmatched API path answers with the
		// JSON envelope instead of the HTML 404 page. They only take effect
		// once the group has at least one route: a chi sub-router with none
		// never builds its middleware chain and answers straight from the
		// fallback, skipping auth and CSRF.
		api.NotFound(s.handlers.APINotFound)
		api.MethodNotAllowed(s.handlers.APIMethodNotAllowed)

		api.Get("/session", s.handlers.GetAPISession)
		api.Get("/categories", s.handlers.GetAPICategories)

		api.Route("/recurrent-expenses", func(recurrentExpenses chi.Router) {
			recurrentExpenses.Get("/", s.handlers.GetAPIRecurrentExpenses)
			recurrentExpenses.Post("/", s.handlers.PostAPIRecurrentExpenses)
			recurrentExpenses.Route("/{id}", func(recurrentExpenses chi.Router) {
				recurrentExpenses.Use(s.handlers.APIRecurrentExpenseContext)

				recurrentExpenses.Get("/", s.handlers.GetAPIRecurrentExpense)
				recurrentExpenses.Put("/", s.handlers.PutAPIRecurrentExpense)
				recurrentExpenses.Delete("/", s.handlers.DeleteAPIRecurrentExpense)
				recurrentExpenses.Post("/unarchive", s.handlers.PostAPIRecurrentExpenseUnarchive)
			})
		})
	})
}

// staticCacheControl lets the browser reuse assets across page loads without a
// request at all. The bundle filenames are content-hashed (manifest.go), so a
// deploy can no longer strand a stale bundle here; the stylesheet and the
// images are not, which is why the window stays short. Once it lapses
// http.FileServer answers the revalidation with a 304 off Last-Modified.
const staticCacheControl = "public, max-age=300"

// setUpFileServer mounts the assets on the root router, outside the app
// middleware chain. Serving a file must not cost a session load, a CSRF token
// or a current-user query.
func (s *Server) setUpFileServer() {
	fileServer := http.FileServer(http.Dir("./web/static/"))

	s.Router.Handle("/static/*", staticCacheHeaders(http.StripPrefix("/static/", fileServer)))
}

func staticCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", staticCacheControl)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) setUpSession() {
	s.Session.Lifetime = 7 * 24 * time.Hour
	s.Session.Cookie.Secure = s.app.IsProduction()
	s.Session.Cookie.HttpOnly = true
	s.Session.Cookie.Persist = true
	s.Session.Cookie.SameSite = http.SameSiteLaxMode
	s.Session.Cookie.Name = "ninete_session"
}

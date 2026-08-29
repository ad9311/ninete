package serve

import (
	"net/http"
	"time"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func (s *Server) setUpRoutes() {
	s.setUpFileServer()
	s.setUpAPIRoutes()

	s.Router.Group(func(root chi.Router) {
		s.setUpAppMiddlewares(root)

		root.Post(cspReportPath, s.handlers.PostCSPReport)
		root.Post("/logout", s.handlers.PostLogout)

		// The expense export. It answers with a file rather than a page, but
		// it belongs to this chain and not to /api: it is reached by a plain
		// anchor, so an expired session has to produce a redirect the browser
		// can follow. Registered before the "/*" catch-all, which would
		// otherwise serve the SPA shell for it.
		root.Get(handlers.ExportExpensesPath, s.handlers.GetExportsExpenses)

		// The SPA shell (docs/spa-migration.md, Phase 7). Wildcarded so every
		// client route resolves on a hard refresh, including one the client
		// router itself does not recognize — that is its own "not found" to
		// show, not the server's. AuthMiddleware guards it like any other
		// page, except for "/login" and "/register", which it exempts.
		root.Get("/", s.handlers.GetApp)
		root.Get("/*", s.handlers.GetApp)
	})
}

// setUpAPIRoutes mounts the JSON API the SPA talks to. It is a sibling of the
// page group, not a child: the two chains differ (see setUpAPIMiddlewares), and
// an /api route must never fall through to a rendered template.
func (s *Server) setUpAPIRoutes() {
	s.Router.Route(apiPathPrefix, func(api chi.Router) {
		s.setUpAPIMiddlewares(api)

		// Registered on the group so an unmatched API path answers with the
		// JSON envelope instead of the HTML 404 page. They only take effect
		// once the group has at least one route: a chi sub-router with none
		// never builds its middleware chain and answers straight from the
		// fallback, skipping auth and CSRF.
		api.NotFound(s.handlers.APINotFound)
		api.MethodNotAllowed(s.handlers.APIMethodNotAllowed)

		// One value for both credential routes (the CLAUDE.md invariant).
		// authRateLimit builds a counter per call, so calling it once per
		// route would hand a client twice the allowance.
		credentialLimit := s.authRateLimit()
		api.With(credentialLimit).Post("/login", s.handlers.PostAPILogin)
		api.With(credentialLimit).Post("/register", s.handlers.PostAPIRegister)

		api.Get("/session", s.handlers.GetAPISession)
		api.Get("/categories", s.handlers.GetAPICategories)
		api.Get("/dashboard", s.handlers.GetAPIDashboard)

		api.Route("/delete-data", func(deleteData chi.Router) {
			deleteData.Get("/", s.handlers.GetAPIDeleteData)
			deleteData.Delete("/", s.handlers.DeleteAPIDeleteDataAll)
			deleteData.Delete("/expenses", s.handlers.DeleteAPIDeleteDataExpenses)
			deleteData.Delete("/recurrent-expenses", s.handlers.DeleteAPIDeleteDataRecurrentExpenses)
			deleteData.Delete("/expense-budgets", s.handlers.DeleteAPIDeleteDataExpenseBudgets)
			deleteData.Delete("/tags", s.handlers.DeleteAPIDeleteDataTags)
		})

		api.Route("/expenses", func(expenses chi.Router) {
			expenses.Get("/", s.handlers.GetAPIExpenses)
			expenses.Post("/", s.handlers.PostAPIExpenses)
			expenses.Post("/quick", s.handlers.PostAPIExpensesQuick)
			expenses.Get("/stats", s.handlers.GetAPIExpensesStats)
			expenses.Get("/budgets", s.handlers.GetAPIExpenseBudgets)
			expenses.Put("/budgets", s.handlers.PutAPIExpenseBudgets)
			expenses.Route("/{id}", func(expenses chi.Router) {
				expenses.Use(s.handlers.APIExpenseContext)

				expenses.Get("/", s.handlers.GetAPIExpense)
				expenses.Put("/", s.handlers.PutAPIExpense)
				expenses.Delete("/", s.handlers.DeleteAPIExpense)
			})
		})

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

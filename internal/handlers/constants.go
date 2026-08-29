package handlers

// ContextKey is used for request context keys managed by HTTP middleware.
type ContextKey string

const (
	KeyCurrentUser      = ContextKey("userID")
	KeyTemplateData     = ContextKey("templateData")
	KeyCSPNonce         = ContextKey("cspNonce")
	KeyExpense          = ContextKey("expenseID")
	KeyRecurrentExpense = ContextKey("recurrentExpenseID")

	// Session keys used in the session store for auth state.
	SessionIsUserSignedIn = "isUserSignedIn"
	SessionUserID         = "userID"
)

// Where a redirect puts a person who has to be sent somewhere to sign in, or
// who is already signed in and has no business on a guest page. Phase 7 of
// docs/spa-migration.md moved the SPA from /app/* to "/", so these carry no
// prefix any more.
//
// lib/api.ts holds AppLoginPath as its own literal for the 401 case, since a
// Go constant cannot reach the bundle. The two must agree.
// AppDashboardPath is "/", not "/dashboard": router.ts maps the dashboard to
// "/", so a redirect to "/dashboard" would render App.svelte's "Not found."
const (
	AppLoginPath     = "/login"
	AppDashboardPath = "/"

	// ExportExpensesPath is on the page chain rather than under /api so an
	// expired session redirects to the login page instead of answering a 401
	// the browser cannot act on (see GetExportsExpenses). Like AppLoginPath, a
	// bundle cannot import a Go constant: routes/exports/Index.svelte holds
	// this literal a second time, and the two have to change together.
	ExportExpensesPath = "/exports/expenses.json"
)

// -------------------------------------------------------------- //

// TemplateName identifies a template by its `<domain>/<view>` path.
type TemplateName string

// AppIndex is the SPA shell, and the only template left once Phase 7 of
// docs/spa-migration.md deleted every rendered page. It carries its own
// <html> document rather than a shared "layout" chrome, since the Svelte app
// owns header/footer/nav.
const AppIndex TemplateName = "app/index"

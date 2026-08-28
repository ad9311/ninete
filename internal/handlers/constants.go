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
// who is already signed in and has no business on a guest page. Phase 6 of
// docs/spa-migration.md ported the auth views, so these point at the SPA: the
// templates still answer /login and /dashboard, but nothing routes a person to
// them any more. Both lose the prefix in Phase 7, when the SPA moves to "/".
//
// lib/api.ts holds AppLoginPath as its own literal for the 401 case, since a
// Go constant cannot reach the bundle. The two must agree.
// AppDashboardPath is BASE_PATH itself, not "/app/dashboard": router.ts maps
// the dashboard to "/", so the SPA has no "/dashboard" route and a redirect
// there renders App.svelte's "Not found." The template UI keeps /dashboard.
const (
	AppLoginPath     = "/app/login"
	AppDashboardPath = "/app"
)

// -------------------------------------------------------------- //

// TemplateName identifies a template by its `<domain>/<view>` path.
type TemplateName string

const (
	// Account templates.
	AccountIndex TemplateName = "account/index"

	// Delete data templates.
	DeleteDataIndex TemplateName = "delete_data/index"

	// Dashboard templates.
	DashboardIndex TemplateName = "dashboard/index"

	// Exports templates.
	ExportsIndex TemplateName = "exports/index"

	// Auth templates.
	LoginIndex    TemplateName = "login/index"
	RegisterIndex TemplateName = "register/index"

	// Expense templates.
	ExpensesIndex   TemplateName = "expenses/index"
	ExpensesNew     TemplateName = "expenses/new"
	ExpensesEdit    TemplateName = "expenses/edit"
	ExpensesShow    TemplateName = "expenses/show"
	ExpensesStats   TemplateName = "expenses/stats"
	ExpensesBudgets TemplateName = "expenses/budgets"

	// Recurrent expense templates.
	RecurrentExpensesIndex TemplateName = "recurrent_expenses/index"
	RecurrentExpensesNew   TemplateName = "recurrent_expenses/new"
	RecurrentExpensesEdit  TemplateName = "recurrent_expenses/edit"
	RecurrentExpensesShow  TemplateName = "recurrent_expenses/show"

	RecurrentExpensesArchived TemplateName = "recurrent_expenses/archived"

	// System templates.
	ErrorIndex    TemplateName = "error/index"
	NotFoundIndex TemplateName = "not_found/index"

	// SPA shell. Served under /app/* until Phase 7 of docs/spa-migration.md
	// moves it to "/"; the template carries its own <html> document rather than
	// the shared "layout" chrome, since the Svelte app owns header/footer/nav.
	AppIndex TemplateName = "app/index"
)

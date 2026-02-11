package handlers

// ContextKey is used for request context keys managed by HTTP middleware.
type ContextKey string

const (
	TemplateData = ContextKey("templateData")

	// Session keys used in the session store for auth state.
	SessionIsUserSignedIn = "isUserSignedIn"
	SessionUserID         = "userID"
)

// -------------------------------------------------------------- //

// TemplateName identifies a template by its `<domain>/<view>` path.
type TemplateName string

const (
	// Dashboard templates.
	DashboardIndex TemplateName = "dashboard/index"

	// Auth templates.
	LoginIndex TemplateName = "login/index"

	// Expense templates.
	ExpensesIndex TemplateName = "expenses/index"
	ExpensesNew   TemplateName = "expenses/new"

	// System templates.
	ErrorIndex    TemplateName = "error/index"
	NotFoundIndex TemplateName = "not_found/index"
)

package handlers

// ContextKey is used for request context keys managed by HTTP middleware.
type ContextKey string

const (
	TemplateData = ContextKey("templateData")
)

// Session keys used in the session store for auth state.
const (
	SessionIsUserSignedIn = "isUserSignedIn"
	SessionUserID         = "userID"
)

// TemplateName identifies a template by its `<domain>/<view>` path.
type TemplateName string

// Dashboard templates.
const (
	DashboardIndex TemplateName = "dashboard/index"
)

// Auth templates.
const (
	LoginIndex TemplateName = "login/index"
)

// Expense templates.
const (
	ExpensesIndex TemplateName = "expenses/index"
)

// System templates.
const (
	ErrorIndex    TemplateName = "error/index"
	NotFoundIndex TemplateName = "not_found/index"
)

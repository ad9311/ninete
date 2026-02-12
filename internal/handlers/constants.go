package handlers

// ContextKey is used for request context keys managed by HTTP middleware.
type ContextKey string

const (
	KeyCurrentUser      = ContextKey("userID")
	KeyTemplateData     = ContextKey("templateData")
	KeyExpense          = ContextKey("expenseID")
	KeyRecurrentExpense = ContextKey("recurrentExpenseID")

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
	ExpensesEdit  TemplateName = "expenses/edit"

	// Recurrent expense templates.
	RecurrentExpensesIndex TemplateName = "recurrent_expenses/index"
	RecurrentExpensesNew   TemplateName = "recurrent_expenses/new"
	RecurrentExpensesEdit  TemplateName = "recurrent_expenses/edit"

	// System templates.
	ErrorIndex    TemplateName = "error/index"
	NotFoundIndex TemplateName = "not_found/index"
)

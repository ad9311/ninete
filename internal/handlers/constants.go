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
)

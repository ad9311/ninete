package handlers

// ContextKey is used for request context keys managed by HTTP middleware.
type ContextKey string

const (
	KeyCurrentUser      = ContextKey("userID")
	KeyTemplateData     = ContextKey("templateData")
	KeyExpense          = ContextKey("expenseID")
	KeyRecurrentExpense = ContextKey("recurrentExpenseID")
	KeyList             = ContextKey("listID")
	KeyTask             = ContextKey("taskID")
	KeyMacroEntry       = ContextKey("macroEntryID")
	KeyMacroTemplate    = ContextKey("macroTemplateID")
	KeyFood             = ContextKey("foodID")

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

	// Exports templates.
	ExportsIndex TemplateName = "exports/index"

	// Auth templates.
	LoginIndex    TemplateName = "login/index"
	RegisterIndex TemplateName = "register/index"

	// Expense templates.
	ExpensesIndex TemplateName = "expenses/index"
	ExpensesNew   TemplateName = "expenses/new"
	ExpensesEdit  TemplateName = "expenses/edit"
	ExpensesShow  TemplateName = "expenses/show"
	ExpensesStats TemplateName = "expenses/stats"

	// Recurrent expense templates.
	RecurrentExpensesIndex TemplateName = "recurrent_expenses/index"
	RecurrentExpensesNew   TemplateName = "recurrent_expenses/new"
	RecurrentExpensesEdit  TemplateName = "recurrent_expenses/edit"
	RecurrentExpensesShow  TemplateName = "recurrent_expenses/show"

	// List templates.
	ListsIndex TemplateName = "lists/index"
	ListsNew   TemplateName = "lists/new"
	ListsEdit  TemplateName = "lists/edit"
	ListsShow  TemplateName = "lists/show"

	// Task templates.
	TasksNew  TemplateName = "tasks/new"
	TasksEdit TemplateName = "tasks/edit"
	TasksShow TemplateName = "tasks/show"

	// Macro templates.
	MacrosIndex TemplateName = "macros/index"
	MacrosNew   TemplateName = "macros/new"
	MacrosEdit  TemplateName = "macros/edit"
	MacrosShow  TemplateName = "macros/show"
	MacrosGoals TemplateName = "macros/goals"
	MacrosStats TemplateName = "macros/stats"

	// Macro template templates.
	MacroTemplatesIndex TemplateName = "macro_templates/index"
	MacroTemplatesNew   TemplateName = "macro_templates/new"
	MacroTemplatesEdit  TemplateName = "macro_templates/edit"
	MacroTemplatesShow  TemplateName = "macro_templates/show"

	// Food templates.
	FoodsIndex TemplateName = "foods/index"
	FoodsNew   TemplateName = "foods/new"
	FoodsEdit  TemplateName = "foods/edit"
	FoodsShow  TemplateName = "foods/show"

	// System templates.
	ErrorIndex    TemplateName = "error/index"
	NotFoundIndex TemplateName = "not_found/index"
)

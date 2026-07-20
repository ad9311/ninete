package handlers

// ContextKey is used for request context keys managed by HTTP middleware.
type ContextKey string

const (
	KeyCurrentUser      = ContextKey("userID")
	KeyTemplateData     = ContextKey("templateData")
	KeyCSPNonce         = ContextKey("cspNonce")
	KeyExpense          = ContextKey("expenseID")
	KeyRecurrentExpense = ContextKey("recurrentExpenseID")
	KeyMacroEntry       = ContextKey("macroEntryID")
	KeyFood             = ContextKey("foodID")
	KeyMoodEntry        = ContextKey("moodEntryID")

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

	// Macro templates.
	MacrosIndex TemplateName = "macros/index"
	MacrosNew   TemplateName = "macros/new"
	MacrosEdit  TemplateName = "macros/edit"
	MacrosShow  TemplateName = "macros/show"
	MacrosGoals TemplateName = "macros/goals"
	MacrosStats TemplateName = "macros/stats"

	// Food templates.
	FoodsIndex TemplateName = "foods/index"
	FoodsNew   TemplateName = "foods/new"
	FoodsEdit  TemplateName = "foods/edit"
	FoodsShow  TemplateName = "foods/show"

	// Mood entry templates.
	MoodEntriesIndex TemplateName = "mood_entries/index"
	MoodEntriesNew   TemplateName = "mood_entries/new"
	MoodEntriesEdit  TemplateName = "mood_entries/edit"
	MoodEntriesShow  TemplateName = "mood_entries/show"
	MoodEntriesStats TemplateName = "mood_entries/stats"

	// System templates.
	ErrorIndex    TemplateName = "error/index"
	NotFoundIndex TemplateName = "not_found/index"
)

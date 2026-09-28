package api

// The server's JSON shapes, copied rather than imported (see the package
// comment). Amounts are cents, "date" is a billed month as UTC midnight of the
// 1st, and created_at is an instant — all in epoch seconds.

// ListedExpense is an expense as GET /api/expenses lists it. The list never
// carries the note; Expense adds it for the single-expense endpoints. They are
// two types so the contract check (docs/mcp.md) can tell that the list really
// does not send a note, instead of this module expecting one it never gets.
type ListedExpense struct {
	ID           int      `json:"id"`
	CategoryID   int      `json:"category_id"`
	CategoryName string   `json:"category_name"`
	Description  string   `json:"description"`
	Amount       uint64   `json:"amount"`
	Date         int64    `json:"date"`
	CreatedAt    int64    `json:"created_at"`
	Tags         []string `json:"tags"`
}

// Expense is the single-expense view: GET, POST and PUT on /api/expenses.
type Expense struct {
	ListedExpense

	Note string `json:"note"`
}

type ExpenseBody struct {
	CategoryID  int      `json:"category_id"`
	Description string   `json:"description"`
	Amount      uint64   `json:"amount"`
	Date        int64    `json:"date"`
	Note        string   `json:"note"`
	Tags        []string `json:"tags"`
}

type QuickExpenseBody struct {
	QuickInput string `json:"quick_input"`
	CategoryID int    `json:"category_id,omitempty"`
	TZOffset   int    `json:"tz_offset"`
}

type Pagination struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	PerPage     int  `json:"per_page"`
	TotalCount  int  `json:"total_count"`
	HasNext     bool `json:"has_next"`
}

type ExpenseList struct {
	Data       []ListedExpense `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

type RecurrentExpense struct {
	ID              int      `json:"id"`
	CategoryID      int      `json:"category_id"`
	CategoryName    string   `json:"category_name"`
	Description     string   `json:"description"`
	Amount          uint64   `json:"amount"`
	Period          uint     `json:"period"`
	OccurrenceLimit uint     `json:"occurrence_limit"`
	OccurrenceCount uint     `json:"occurrence_count"`
	Archived        bool     `json:"archived"`
	Tags            []string `json:"tags"`
}

type RecurrentExpenseBody struct {
	CategoryID      int      `json:"category_id"`
	Description     string   `json:"description"`
	Amount          uint64   `json:"amount"`
	Period          uint     `json:"period"`
	OccurrenceLimit uint     `json:"occurrence_limit"`
	Tags            []string `json:"tags"`
}

type RecurrentExpenseList struct {
	Data       []RecurrentExpense `json:"data"`
	Pagination Pagination         `json:"pagination"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CategoryList struct {
	Data []Category `json:"data"`
}

type CategoryTotal struct {
	Name  string `json:"name"`
	Total uint64 `json:"total"`
}

type ExpenseStats struct {
	Data []CategoryTotal `json:"data"`
}

type Dashboard struct {
	Data struct {
		ThisMonthTotal  uint64          `json:"this_month_total"`
		LastMonthTotal  uint64          `json:"last_month_total"`
		MonthChangeSign string          `json:"month_change_sign"`
		MonthChangePct  int             `json:"month_change_pct"`
		TopCategories   []CategoryTotal `json:"top_categories"`
	} `json:"data"`
}

type BudgetMonth struct {
	Month string `json:"month"`
	Total uint64 `json:"total"`
	Pct   int    `json:"pct"`
	Over  bool   `json:"over"`
}

type BudgetRow struct {
	CategoryName string        `json:"category_name"`
	Total        uint64        `json:"total"`
	HasBudget    bool          `json:"has_budget"`
	Budget       uint64        `json:"budget"`
	Left         int64         `json:"left"`
	Pct          int           `json:"pct"`
	Over         bool          `json:"over"`
	Months       []BudgetMonth `json:"months"`
	MonthsOver   int           `json:"months_over"`
	MonthCount   int           `json:"month_count"`
	AvgPerMonth  uint64        `json:"avg_per_month"`
}

type BudgetEditRow struct {
	CategoryID int    `json:"category_id"`
	Name       string `json:"name"`
	Amount     uint64 `json:"amount"`
}

type Budgets struct {
	Mode     string          `json:"mode"`
	Rows     []BudgetRow     `json:"rows"`
	EditRows []BudgetEditRow `json:"edit_rows"`
}

type BudgetsBody struct {
	// Amounts is keyed by category id as a string. A zero amount clears that
	// category's budget; a category left out is untouched.
	Amounts map[string]uint64 `json:"amounts"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ReportSettings struct {
	SelectedTagIDs []int `json:"selected_tag_ids"`
	Tags           []Tag `json:"tags"`
	TagLimit       int   `json:"tag_limit"`
}

type ReportSettingsBody struct {
	TagIDs []int `json:"tag_ids"`
}

type RetagBody struct {
	From []string `json:"from"`
	To   string   `json:"to"`
}

type Retag struct {
	Tag      Tag `json:"tag"`
	Retagged int `json:"retagged"`
}

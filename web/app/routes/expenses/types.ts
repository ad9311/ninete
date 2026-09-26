// Mirrors handle_api_expenses.go's, handle_api_expense_stats.go's and
// handle_api_expense_budgets.go's JSON shapes exactly, snake_case included
// (§3.5 of docs/spa-migration.md), so nothing maps between the wire format and
// what a component reads. date is a calendar date (UTC midnight), created_at
// an instant — see §3.6 before formatting either.

export interface Expense {
  id: number;
  category_id: number;
  category_name: string;
  description: string;
  amount: number;
  date: number;
  created_at: number;
  tags: string[];
}

// The single-expense shape (GET /api/expenses/{id}, and the POST/PUT
// responses). The note is deliberately absent from Expense: the list endpoint
// does not send it, since it is shown only on the expense's own page.
export interface ExpenseDetail extends Expense {
  note: string;
}

// Mirrors ExpenseParams.Note's max=255 in logic_expense.go. The server counts
// characters (runes) after trimming, which is what noteLength counts too.
export const NOTE_MAX_LENGTH = 255;

export function noteLength(note: string): number {
  return [...note.trim()].length;
}

export interface Pagination {
  current_page: number;
  total_pages: number;
  per_page: number;
  total_count: number;
  has_prev: boolean;
  has_next: boolean;
  sort_field: string;
  sort_order: string;
  category_id: number;
}

export interface ExpenseListResponse {
  data: Expense[];
  pagination: Pagination;
}

export interface ExpenseRequestBody {
  category_id: number;
  description: string;
  amount: number;
  date: number;
  note: string;
  tags: string[];
}

export interface QuickExpenseRequestBody {
  quick_input: string;
  category_id?: number;
  tz_offset: number;
}

export interface ExpenseStatRow {
  name: string;
  total: number;
}

export interface ExpenseStatsResponse {
  data: ExpenseStatRow[];
}

export type BudgetMode = "month" | "months";

export interface BudgetMonthRow {
  month: string;
  total: number;
  pct: number;
  bar_pct: number;
  over: boolean;
}

export interface BudgetRow {
  category_name: string;
  total: number;
  has_budget: boolean;
  budget: number;
  left: number;
  pct: number;
  bar_pct: number;
  over: boolean;
  months: BudgetMonthRow[];
  months_over: number;
  month_count: number;
  avg_per_month: number;
}

export interface BudgetEditRow {
  category_id: number;
  name: string;
  amount: number;
}

export interface ExpenseBudgetsResponse {
  mode: BudgetMode;
  rows: BudgetRow[];
  edit_rows: BudgetEditRow[];
}

// Mirrors internal/handlers/handle_api_recurrent_expenses.go's JSON shape
// exactly — snake_case field names (§3.5 of docs/spa-migration.md), so no
// mapping layer sits between the wire format and what components read.

export interface RecurrentExpense {
  id: number;
  category_id: number;
  category_name: string;
  description: string;
  amount: number;
  period: number;
  occurrence_limit: number;
  occurrence_count: number;
  archived: boolean;
  tags: string[];
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

export interface RecurrentExpenseListResponse {
  data: RecurrentExpense[];
  pagination: Pagination;
}

export interface RecurrentExpenseRequestBody {
  category_id: number;
  description: string;
  amount: number;
  period: number;
  occurrence_limit: number;
  tags: string[];
}

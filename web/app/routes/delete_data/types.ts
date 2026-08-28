export interface DeleteDataCounts {
  expenses: number;
  recurrent_expenses: number;
  expense_budgets: number;
  tags: number;
}

export interface DeleteDataCountsResponse {
  data: DeleteDataCounts;
}

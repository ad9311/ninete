// Mirrors handle_api_dashboard.go's JSON shape exactly (§3.5).

export interface DashboardCategoryTotal {
  name: string;
  total: number;
}

export interface DashboardSummary {
  this_month_total: number;
  last_month_total: number;
  month_change_sign: string;
  month_change_pct: number;
  top_categories: DashboardCategoryTotal[];
}

export interface DashboardResponse {
  data: DashboardSummary;
}

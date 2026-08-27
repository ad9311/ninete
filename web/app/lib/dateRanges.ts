// Client-side twin of computeDateRange (internal/handlers/expense_shared.go).
// §3.6 of docs/spa-migration.md retires tz_offset on the API side for named
// ranges: the client already knows its own zone (the browser's local Date
// getters), so it resolves a range key to explicit UTC-midnight [start, end)
// epoch-second bounds itself, instead of sending the key and an offset for the
// server to resolve. Every branch below mirrors computeDateRange's Go
// arithmetic one for one — do not hand-simplify the month math, since
// Date.UTC's overflow normalization is what makes that mirroring exact.

export interface DateRangeOption {
  value: string;
  label: string;
}

/** Matches handlers.dateRangeLabels — the expenses/stats filter options. */
export const DATE_RANGE_OPTIONS: DateRangeOption[] = [
  { value: "this_month", label: "This month" },
  { value: "next_month", label: "Next month" },
  { value: "last_month", label: "Last month" },
  { value: "this_week", label: "This week" },
  { value: "six_months", label: "Last 6 months" },
  { value: "this_year", label: "This year" },
];

export type BudgetMode = "month" | "months";

export interface BudgetDateRangeOption extends DateRangeOption {
  mode: BudgetMode;
}

/** Matches handlers.budgetDateRanges — the budgets page's range/mode pairs. */
export const BUDGET_DATE_RANGE_OPTIONS: BudgetDateRangeOption[] = [
  { value: "this_month", label: "This month", mode: "month" },
  { value: "last_month", label: "Last month", mode: "month" },
  { value: "six_months", label: "Last 6 months", mode: "months" },
  { value: "this_year", label: "This year", mode: "months" },
];

export interface DateBounds {
  start: number;
  end: number;
}

// Date.UTC normalizes an out-of-range month (13, -1, ...) the same way Go's
// time.Date normalizes an out-of-range time.Month, so month arithmetic can
// carry values outside 0-11 and land on the right year.
function startOfMonthUTC(year: number, month: number): number {
  return Date.UTC(year, month, 1, 0, 0, 0, 0) / 1000;
}

/**
 * Resolves a named range key to [start, end) UTC-midnight epoch-second
 * bounds, using the browser's local calendar to decide what "now" means —
 * the same role tz_offset played server-side. Returns null for "all_time" or
 * an unrecognized key, meaning "no date filter".
 */
export function computeDateRange(key: string): DateBounds | null {
  const now = new Date();
  const year = now.getFullYear();
  const month = now.getMonth();

  switch (key) {
    case "this_month":
      return {
        start: startOfMonthUTC(year, month),
        end: startOfMonthUTC(year, month + 1),
      };
    case "next_month":
      return {
        start: startOfMonthUTC(year, month + 1),
        end: startOfMonthUTC(year, month + 2),
      };
    case "last_month":
      return {
        start: startOfMonthUTC(year, month - 1),
        end: startOfMonthUTC(year, month),
      };
    case "this_week": {
      // Go: weekday := now.Weekday() (Sunday=0); if Sunday, treat as 7, so
      // Monday=1..Sunday=7, then step back to Monday. JS's getDay() is the
      // same Sunday=0 scheme, so the mirroring is exact.
      let weekday = now.getDay();
      if (weekday === 0) weekday = 7;
      const monday = new Date(now);
      monday.setDate(now.getDate() - (weekday - 1));
      const start =
        Date.UTC(monday.getFullYear(), monday.getMonth(), monday.getDate()) /
        1000;

      return { start, end: start + 7 * 24 * 60 * 60 };
    }
    case "six_months":
      // Five months back plus the current one is six, matching the label.
      return {
        start: startOfMonthUTC(year, month - 5),
        end: startOfMonthUTC(year, month + 1),
      };
    case "this_year":
      return {
        start: Date.UTC(year, 0, 1) / 1000,
        end: Date.UTC(year + 1, 0, 1) / 1000,
      };
    default:
      return null;
  }
}

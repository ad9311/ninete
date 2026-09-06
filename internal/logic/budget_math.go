package logic

// The budget arithmetic, here rather than in internal/handlers because two
// callers need it: the budgets endpoint that has always rendered it, and the
// monthly report's budget table. A second copy is exactly the kind that drifts
// — the report would keep saying "under budget" after the endpoint's rule
// changed.

// BudgetLeft is what remains of a budget, negative once it is exceeded. Both
// operands are cent amounts one person entered by hand, so neither half of the
// subtraction can approach the int64 range.
func BudgetLeft(budget, total uint64) int64 {
	if budget >= total {
		return int64(budget - total) //nolint:gosec // cent amount, far below int64 max
	}

	return -int64(total - budget) //nolint:gosec // cent amount, far below int64 max
}

// BudgetPercent returns the true percent and the percent clamped to 100 for a
// progress bar. A zero budget never reaches here — a cleared amount is deleted
// rather than stored — but it is guarded anyway, since it would divide.
func BudgetPercent(total, budget uint64) (int, int) {
	if budget == 0 {
		return 0, 0
	}

	pct := int(total * 100 / budget) //nolint:gosec // both operands are page-sized cent amounts
	if pct > 100 {
		return pct, 100
	}

	return pct, pct
}

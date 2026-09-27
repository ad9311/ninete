package tools

import "errors"

var (
	ErrTagsConflict  = errors.New("pass either tags (to replace the whole list) or add_tags/remove_tags, not both")
	ErrNothingToSet  = errors.New("nothing to change: pass at least one field to update")
	ErrPerPage       = errors.New("per_page must be 15, 25, 50 or 100")
	ErrSortField     = errors.New("unsupported sort_field")
	ErrSortOrder     = errors.New(`sort_order must be "asc" or "desc"`)
	ErrRangeConflict = errors.New(
		"filter by billed month (from_month/to_month) or by creation day (created_from/created_to), not both",
	)
	ErrToWithoutFrom = errors.New("to_month or created_to needs its from_month or created_from")
	ErrUnknownTags   = errors.New("unknown tags (tags are created by adding them to an expense)")
	ErrNoBudgets     = errors.New("budgets must list at least one category")
	ErrBudgetMode    = errors.New(`mode must be "month" or "months"`)
)

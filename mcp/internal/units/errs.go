package units

import "errors"

var (
	ErrAmount = errors.New(`amount must be a positive decimal with at most two decimals, like "12.50"`)
	ErrMonth  = errors.New(`month must be in YYYY-MM format, like "2026-09"`)
	ErrDay    = errors.New(`date must be in YYYY-MM-DD format, like "2026-09-26"`)
	ErrRange  = errors.New("the range ends before it starts")
)

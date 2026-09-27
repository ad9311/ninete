// Package units converts between the values a person (or a model) writes and
// the values the Ninete API speaks.
//
// The API's rules, which every function here mirrors:
//   - Money is unsigned integer cents. Never a float.
//   - An expense's billed date is a calendar *month*, stored as UTC midnight of
//     the 1st. It carries no zone, so converting it needs none.
//   - created_at is an instant. Turning a day the user names into a window of
//     instants needs a zone, which the client supplies (the browser's; here,
//     NINETE_TZ). The server never guesses one.
package units

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrAmount = errors.New(`amount must be a positive decimal with at most two decimals, like "12.50"`)
	ErrMonth  = errors.New(`month must be in YYYY-MM format, like "2026-09"`)
	ErrDay    = errors.New(`date must be in YYYY-MM-DD format, like "2026-09-26"`)
	ErrRange  = errors.New("the range ends before it starts")
)

const (
	monthLayout = "2006-01"
	dayLayout   = "2006-01-02"
	centsPerOne = 100
)

// ParseAmount turns "12.50", "12.5" or "12" into cents. Zero is allowed —
// SetBudgets uses it to clear a budget — and callers that need a positive
// amount leave that to the API's own validation.
func ParseAmount(raw string) (uint64, error) {
	raw = strings.TrimSpace(raw)

	whole, frac, hasFrac := strings.Cut(raw, ".")
	if whole == "" || (hasFrac && (frac == "" || len(frac) > 2)) {
		return 0, ErrAmount
	}

	for _, part := range []string{whole, frac} {
		for _, r := range part {
			if r < '0' || r > '9' {
				return 0, ErrAmount
			}
		}
	}

	units, err := strconv.ParseUint(whole, 10, 64)
	if err != nil {
		return 0, ErrAmount
	}

	var cents uint64
	if hasFrac {
		if len(frac) == 1 {
			frac += "0"
		}

		cents, err = strconv.ParseUint(frac, 10, 64)
		if err != nil {
			return 0, ErrAmount
		}
	}

	return units*centsPerOne + cents, nil
}

// FormatAmount turns cents into "12.50". A string, not a float, so a total is
// never shown as 12.499999.
func FormatAmount(cents uint64) string {
	return fmt.Sprintf("%d.%02d", cents/centsPerOne, cents%centsPerOne)
}

// FormatSignedAmount is FormatAmount for the budget "left" figure, which goes
// negative when a category is over budget.
func FormatSignedAmount(cents int64) string {
	if cents < 0 {
		return "-" + FormatAmount(uint64(-cents))
	}

	return FormatAmount(uint64(cents))
}

// ParseMonth turns "2026-09" into the billed date the API stores for it: UTC
// midnight of the 1st.
func ParseMonth(raw string) (int64, error) {
	t, err := time.Parse(monthLayout, strings.TrimSpace(raw))
	if err != nil {
		return 0, ErrMonth
	}

	return t.Unix(), nil
}

// FormatMonth turns a stored billed date back into "2026-09". UTC getters on
// purpose: the value is a calendar date, and reading it in a local zone west
// of UTC would move it into the previous month.
func FormatMonth(unix int64) string {
	return time.Unix(unix, 0).UTC().Format(monthLayout)
}

// MonthRange resolves an inclusive span of billed months to the half-open
// [start, end) bounds the API filters "date" by. An empty to means the single
// month from names.
func MonthRange(from, to string) (start, end int64, err error) {
	if to == "" {
		to = from
	}

	start, err = ParseMonth(from)
	if err != nil {
		return 0, 0, err
	}

	last, err := ParseMonth(to)
	if err != nil {
		return 0, 0, err
	}

	if last < start {
		return 0, 0, ErrRange
	}

	end = time.Unix(last, 0).UTC().AddDate(0, 1, 0).Unix()

	return start, end, nil
}

// DayRange resolves an inclusive span of calendar days, read in loc, to the
// half-open [start, end) instants the API filters created_at by. An empty to
// means the single day from names.
func DayRange(from, to string, loc *time.Location) (start, end int64, err error) {
	if to == "" {
		to = from
	}

	first, err := time.ParseInLocation(dayLayout, strings.TrimSpace(from), loc)
	if err != nil {
		return 0, 0, ErrDay
	}

	last, err := time.ParseInLocation(dayLayout, strings.TrimSpace(to), loc)
	if err != nil {
		return 0, 0, ErrDay
	}

	if last.Before(first) {
		return 0, 0, ErrRange
	}

	// AddDate rather than +24h: a day that crosses a DST change is 23 or 25
	// hours long, and the window must end at the next local midnight.
	return first.Unix(), last.AddDate(0, 0, 1).Unix(), nil
}

// CurrentMonth is the calendar month now falls in, read in loc — the month a
// new expense defaults to, as the web form does with the browser's clock.
func CurrentMonth(now time.Time, loc *time.Location) string {
	return now.In(loc).Format(monthLayout)
}

// PreviousMonth returns the month before a "YYYY-MM" month.
func PreviousMonth(month string) (string, error) {
	t, err := time.Parse(monthLayout, month)
	if err != nil {
		return "", ErrMonth
	}

	return t.AddDate(0, -1, 0).Format(monthLayout), nil
}

// TZOffsetMinutes is what JavaScript's Date.getTimezoneOffset() returns for
// loc at now: minutes *behind* UTC, so UTC-5 is 300. Quick-add takes its zone
// in exactly this form.
func TZOffsetMinutes(now time.Time, loc *time.Location) int {
	_, offsetSeconds := now.In(loc).Zone()

	return -offsetSeconds / 60
}

// FormatInstant renders an instant in loc as RFC 3339, so the zone it was read
// in travels with it.
func FormatInstant(unix int64, loc *time.Location) string {
	return time.Unix(unix, 0).In(loc).Format(time.RFC3339)
}

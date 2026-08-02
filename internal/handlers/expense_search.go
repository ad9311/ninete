package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ad9311/ninete/internal/repo"
)

const (
	// searchDateLayout is the only accepted date input format (YYYY-MM-DD).
	searchDateLayout = "2006-01-02"
	// searchTermMax bounds free-text search input.
	searchTermMax = 50
	secondsPerDay = int64(24 * 60 * 60)
)

// expenseSearch holds the expense index search inputs. The exported fields are
// echoed back to the form; the unexported ones are the parsed date bounds.
type expenseSearch struct {
	Query    string
	Tag      string
	DateFrom string
	DateTo   string

	fromUnix int64
	toUnix   int64
	hasFrom  bool
	hasTo    bool
	// explicitRange records whether the request carried its own date_range, in
	// which case the user's choice wins over any implicit widening.
	explicitRange bool
}

func parseExpenseSearch(r *http.Request) (expenseSearch, error) {
	q := r.URL.Query()
	search := expenseSearch{
		Query:         strings.TrimSpace(q.Get("q")),
		Tag:           strings.TrimSpace(q.Get("tag")),
		DateFrom:      strings.TrimSpace(q.Get("date_from")),
		DateTo:        strings.TrimSpace(q.Get("date_to")),
		explicitRange: q.Get("date_range") != "",
	}

	if utf8.RuneCountInString(search.Query) > searchTermMax ||
		utf8.RuneCountInString(search.Tag) > searchTermMax {
		return search, ErrSearchTermTooLong
	}

	if search.DateFrom != "" {
		from, err := parseSearchDate(search.DateFrom)
		if err != nil {
			return search, err
		}
		search.fromUnix = from
		search.hasFrom = true
	}

	if search.DateTo != "" {
		to, err := parseSearchDate(search.DateTo)
		if err != nil {
			return search, err
		}
		// The bound is inclusive, so filter against the start of the next day.
		search.toUnix = to + secondsPerDay
		search.hasTo = true
	}

	if search.hasFrom && search.hasTo && search.toUnix <= search.fromUnix {
		return search, ErrSearchDateRange
	}

	return search, nil
}

func parseSearchDate(value string) (int64, error) {
	// time.Parse accepts unpadded components such as "2026-1-5"; the form input
	// declares a strict \d{4}-\d{2}-\d{2} pattern, so reject anything the browser
	// itself would refuse rather than echoing it back into an invalid field.
	if len(value) != len(searchDateLayout) {
		return 0, fmt.Errorf("%w: %q", ErrSearchDateFormat, value)
	}

	date, err := time.Parse(searchDateLayout, value)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", ErrSearchDateFormat, value)
	}

	return date.Unix(), nil
}

func (s expenseSearch) hasDateBounds() bool {
	return s.hasFrom || s.hasTo
}

func (s expenseSearch) hasTextSearch() bool {
	return s.Query != "" || s.Tag != ""
}

// clearsPresetRange reports whether the search overrides the preset date-range
// filter. Explicit bounds always win, and a text search implicitly widens to all
// time so that matches older than the default range are not silently hidden —
// unless the request picked a date_range itself.
func (s expenseSearch) clearsPresetRange() bool {
	if s.hasDateBounds() {
		return true
	}

	return s.hasTextSearch() && !s.explicitRange
}

// apply adds the active search predicates to opts, replacing the preset
// date-range filters where the search overrides them.
func (s expenseSearch) apply(opts *repo.QueryOptions, userID int) {
	fields := opts.Filters.FilterFields

	if s.clearsPresetRange() {
		fields = slices.DeleteFunc(fields, func(f repo.FilterField) bool {
			return f.Name == "date"
		})

		if s.hasFrom {
			fields = append(fields, repo.FilterField{Name: "date", Value: s.fromUnix, Operator: ">="})
		}
		if s.hasTo {
			fields = append(fields, repo.FilterField{Name: "date", Value: s.toUnix, Operator: "<"})
		}
	}

	if s.Query != "" {
		fields = append(fields, repo.ExpenseDescriptionFilter(s.Query))
	}

	if s.Tag != "" {
		fields = append(fields, repo.ExpenseTagFilter(userID, s.Tag))
	}

	opts.Filters.FilterFields = fields
}

package handlers

import (
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/ad9311/ninete/internal/repo"
)

const (
	// searchTermMax bounds free-text search input.
	searchTermMax = 50

	// searchFieldBilled is the billed-date column. The explicit search bounds
	// no longer target it — the billed date is picked and displayed as a month,
	// so a day-precision bound on it meant nothing — but the preset date_range
	// filter still does, which is what apply drops it by name for.
	searchFieldBilled = "date"
	// searchFieldCreated is the column the explicit search bounds target.
	searchFieldCreated = "created_at"
)

// expenseSearch holds the expense index search inputs.
//
// The created-date bounds arrive already resolved, as epoch seconds, for the
// same reason named ranges do (§3.6 of docs/spa-migration.md, "Retiring
// tz_offset on the API side"): created_at is an *instant*, so turning the day
// the user typed into a [start, end) window needs their zone, and the client is
// the only party that knows it. The server never sees the YYYY-MM-DD string and
// never has to be told a zone — a fixed server-side zone would be wrong for
// every request made from anywhere else.
type expenseSearch struct {
	Query string
	Tag   string

	createdStart int64
	createdEnd   int64
	hasBounds    bool
	// explicitRange records whether the request carried its own date_range, in
	// which case the user's choice wins over any implicit widening.
	explicitRange bool
}

func parseExpenseSearch(r *http.Request) (expenseSearch, error) {
	q := r.URL.Query()
	search := expenseSearch{
		Query:         strings.TrimSpace(q.Get("q")),
		Tag:           strings.TrimSpace(q.Get("tag")),
		explicitRange: q.Get("date_range") != "",
	}

	if utf8.RuneCountInString(search.Query) > searchTermMax ||
		utf8.RuneCountInString(search.Tag) > searchTermMax {
		return search, ErrSearchTermTooLong
	}

	start, end, hasBounds, err := parseAPICreatedBounds(q)
	if err != nil {
		return search, err
	}
	search.createdStart, search.createdEnd, search.hasBounds = start, end, hasBounds

	return search, nil
}

func (s expenseSearch) hasTextSearch() bool {
	return s.Query != "" || s.Tag != ""
}

// clearsPresetRange reports whether the search overrides the preset date-range
// filter. Explicit bounds always win, and a text search implicitly widens to all
// time so that matches older than the default range are not silently hidden —
// unless the request picked a date_range itself.
func (s expenseSearch) clearsPresetRange() bool {
	if s.hasBounds {
		return true
	}

	return s.hasTextSearch() && !s.explicitRange
}

// apply adds the active search predicates to opts, replacing the preset
// date-range filters where the search overrides them.
func (s expenseSearch) apply(opts *repo.QueryOptions, userID int) {
	fields := opts.Filters.FilterFields

	if s.clearsPresetRange() {
		// The preset range filters on the billed date and the explicit bounds
		// filter on created_at, so the two can never combine into a meaningful
		// window — the preset is dropped whenever the search overrides it.
		fields = slices.DeleteFunc(fields, func(f repo.FilterField) bool {
			return f.Name == searchFieldBilled
		})

		if s.hasBounds {
			fields = append(fields,
				repo.FilterField{
					Name: searchFieldCreated, Value: s.createdStart, Operator: ">=",
				},
				repo.FilterField{
					Name: searchFieldCreated, Value: s.createdEnd, Operator: "<",
				},
			)
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

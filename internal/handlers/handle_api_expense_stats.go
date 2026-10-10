package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/ad9311/ninete/internal/repo"
)

// apiExpenseStatRow's field names double as Chart.js's series shape on the
// client (name/total), so the SPA passes the response straight into the chart
// with no reshaping — see GetExpensesStats's chartPoint, which this mirrors.
type apiExpenseStatRow struct {
	Name  string `json:"name"`
	Total uint64 `json:"total"`
}

type apiExpenseStatsResponse struct {
	Data []apiExpenseStatRow `json:"data"`
}

// GetAPIExpensesStats answers the stats view. It takes explicit [start, end)
// bounds rather than date_range+tz_offset (§3.6 of docs/spa-migration.md).
func (h *Handler) GetAPIExpensesStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	var query apiExpenseStatsQuery
	decodeQuery(r, &query)

	sortField, sortOrder, err := statsSort(query)
	if err != nil {
		h.WriteJSONError(w, http.StatusUnprocessableEntity, err)

		return
	}

	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: user.ID, Operator: "="},
		},
		Connector: "AND",
	}

	start, end, hasBounds, err := parseAPIDateBounds(query.apiBoundsQuery)
	if err != nil {
		h.WriteAPIError(w, err, ErrAPIInvalidDateRange)

		return
	}
	if hasBounds {
		filters.FilterFields = append(filters.FilterFields,
			repo.FilterField{Name: "date", Value: start, Operator: ">="},
			repo.FilterField{Name: "date", Value: end, Operator: "<"},
		)
	}

	totals, err := h.store.FindExpensesCategoryTotals(ctx, filters)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	_, categoryNameByID, err := h.findCategories(ctx)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	rows := make([]apiExpenseStatRow, 0, len(totals))
	for _, t := range totals {
		rows = append(rows, apiExpenseStatRow{
			Name:  categoryNameOrUnknown(categoryNameByID, t.CategoryID),
			Total: t.Total,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		switch sortField {
		case "category":
			if sortOrder == "ASC" {
				return rows[i].Name < rows[j].Name
			}

			return rows[i].Name > rows[j].Name
		default:
			if sortOrder == "ASC" {
				return rows[i].Total < rows[j].Total
			}

			return rows[i].Total > rows[j].Total
		}
	})

	h.WriteJSON(w, http.StatusOK, apiExpenseStatsResponse{Data: rows})
}

// statsSort resolves the stats view's sort, defaulting each half on its own
// (total, DESC) and rejecting a value it does not know. The order is
// case-insensitive, as the repo's sort builder makes it for every list.
func statsSort(query apiExpenseStatsQuery) (field, order string, err error) {
	field = query.SortField
	switch field {
	case "":
		field = "total"
	case "category", "total":
	default:
		return "", "", ErrQueryStatsSortField
	}

	order = strings.ToUpper(query.SortOrder)
	switch order {
	case "":
		order = "DESC"
	case "ASC", "DESC":
	default:
		return "", "", repo.ErrInvalidSortOrder
	}

	return field, order, nil
}

package handlers

import (
	"net/http"
	"sort"

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

// GetAPIExpensesStats is GetExpensesStats's JSON twin: explicit [start, end)
// bounds instead of date_range+tz_offset (§3.6 of docs/spa-migration.md).
func (h *Handler) GetAPIExpensesStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	q := r.URL.Query()

	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: user.ID, Operator: "="},
		},
		Connector: "AND",
	}

	start, end, hasBounds, err := parseAPIDateBounds(q)
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

	sortField := q.Get("sort_field")
	sortOrder := q.Get("sort_order")
	if sortField == "" {
		sortField = "total"
	}
	if sortOrder == "" {
		sortOrder = "DESC"
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

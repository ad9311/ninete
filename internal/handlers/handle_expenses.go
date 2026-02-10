package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/repo"
)

func (h *Handler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	user := getUserContext(r)

	userIDFilter := repo.FilterField{
		Name:     "user_id",
		Value:    user.ID,
		Operator: "=",
	}

	opts := repo.QueryOptions{}
	opts.Filters.FilterFields = append(opts.Filters.FilterFields, userIDFilter)
	opts.Filters.Connector = "AND"

	expenses, err := h.store.FindExpenses(r.Context(), opts)
	if err != nil {
		data["error"] = err.Error()
	}

	data["expenses"] = expenses

	h.render(w, http.StatusOK, ExpensesIndex, data)
}

package handlers

import "net/http"

func (h *Handler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, ExpensesIndex, h.tmplData(r))
}

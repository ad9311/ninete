package handlers

import (
	"net/http"
)

// apiAccountDataCounts is logic.AccountDataCounts as JSON, populating the
// delete-data page's per-section record counts.
type apiAccountDataCounts struct {
	Expenses          int `json:"expenses"`
	RecurrentExpenses int `json:"recurrent_expenses"`
	ExpenseBudgets    int `json:"expense_budgets"`
	Tags              int `json:"tags"`
}

type apiAccountDataCountsResponse struct {
	Data apiAccountDataCounts `json:"data"`
}

func (h *Handler) GetAPIDeleteData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	counts, err := h.store.FindAccountDataCounts(ctx, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.WriteJSON(w, http.StatusOK, apiAccountDataCountsResponse{
		Data: apiAccountDataCounts{
			Expenses:          counts.Expenses,
			RecurrentExpenses: counts.RecurrentExpenses,
			ExpenseBudgets:    counts.ExpenseBudgets,
			Tags:              counts.Tags,
		},
	})
}

func (h *Handler) DeleteAPIDeleteDataExpenses(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)

	if err := h.store.DeleteAllExpenses(r.Context(), user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAPIDeleteDataRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)

	if err := h.store.DeleteAllRecurrentExpenses(r.Context(), user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAPIDeleteDataExpenseBudgets(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)

	if err := h.store.DeleteAllExpenseBudgets(r.Context(), user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAPIDeleteDataTags(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)

	if err := h.store.DeleteAllTags(r.Context(), user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAPIDeleteDataAll(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)

	if err := h.store.DeleteAllUserData(r.Context(), user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

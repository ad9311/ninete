package handlers

import (
	"net/http"
)

func (h *Handler) GetDeleteData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	counts, err := h.store.FindAccountDataCounts(ctx, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	data["counts"] = counts

	h.render(w, http.StatusOK, DeleteDataIndex, data)
}

func (h *Handler) PostDeleteDataExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllExpenses(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account/delete-data", http.StatusSeeOther)
}

func (h *Handler) PostDeleteDataRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllRecurrentExpenses(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account/delete-data", http.StatusSeeOther)
}

func (h *Handler) PostDeleteDataExpenseBudgets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllExpenseBudgets(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account/delete-data", http.StatusSeeOther)
}

func (h *Handler) PostDeleteDataTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllTags(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account/delete-data", http.StatusSeeOther)
}

func (h *Handler) PostDeleteDataAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllUserData(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account/delete-data", http.StatusSeeOther)
}

package handlers

import (
	"net/http"
)

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	counts, err := h.store.FindAccountDataCounts(ctx, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, AccountIndex, err)

		return
	}

	data["counts"] = counts

	h.render(w, http.StatusOK, AccountIndex, data)
}

func (h *Handler) PostAccountDeleteExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllExpenses(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func (h *Handler) PostAccountDeleteRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllRecurrentExpenses(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func (h *Handler) PostAccountDeleteMacroEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllMacroEntries(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func (h *Handler) PostAccountDeleteMacroGoals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllMacroGoals(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func (h *Handler) PostAccountDeleteFoods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllFoods(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func (h *Handler) PostAccountDeleteMoodEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllMoodEntries(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func (h *Handler) PostAccountDeleteTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllTags(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func (h *Handler) PostAccountDeleteAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	if err := h.store.DeleteAllUserData(ctx, user.ID); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/go-chi/chi/v5"
)

// apiTagUsage is one row of the tag page: a tag and how many records of each
// kind carry it. Both counts at zero means the tag is unused.
type apiTagUsage struct {
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	ExpenseCount          int    `json:"expense_count"`
	RecurrentExpenseCount int    `json:"recurrent_expense_count"`
}

type apiTagListResponse struct {
	Data []apiTagUsage `json:"data"`
}

type apiDeleteUnusedTagsResponse struct {
	Deleted int `json:"deleted"`
}

// GetAPITags lists every tag the user owns, sorted by name, unused ones
// included — the only place an orphan left by a retag shows up as one.
func (h *Handler) GetAPITags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	usages, err := h.store.FindTagUsages(ctx, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	data := make([]apiTagUsage, 0, len(usages))
	for _, u := range usages {
		data = append(data, apiTagUsage{
			ID:                    u.ID,
			Name:                  u.Name,
			ExpenseCount:          u.ExpenseCount,
			RecurrentExpenseCount: u.RecurrentExpenseCount,
		})
	}

	h.WriteJSON(w, http.StatusOK, apiTagListResponse{Data: data})
}

// DeleteAPITag removes one tag, in use or not; the records that carried it
// keep everything else. Browser-only in practice: no token scope may DELETE.
func (h *Handler) DeleteAPITag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	id, err := prog.ParseID(chi.URLParam(r, "id"), "Tag")
	if err != nil {
		h.APINotFound(w, r)

		return
	}

	if _, err := h.store.DeleteTag(ctx, id, user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteAPIUnusedTags removes every tag no record carries and says how many
// went. Browser-only in practice, like DeleteAPITag.
func (h *Handler) DeleteAPIUnusedTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	deleted, err := h.store.DeleteUnusedTags(ctx, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.WriteJSON(w, http.StatusOK, apiDeleteUnusedTagsResponse{Deleted: deleted})
}

type retagRequestBody struct {
	From []string `json:"from"`
	To   string   `json:"to"`
}

type apiTag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// apiRetagResponse names the tag the records now carry and how many distinct
// records — expenses and recurrent expenses together — were moved onto it.
type apiRetagResponse struct {
	Tag      apiTag `json:"tag"`
	Retagged int    `json:"retagged"`
}

// PostAPITagsRetag moves every record tagged with any of "from" onto "to" in
// one transaction. The "from" tags are left in place, unused; see
// logic.Store.Retag for why nothing is deleted.
func (h *Handler) PostAPITagsRetag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	var body retagRequestBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	result, err := h.store.Retag(ctx, user.ID, logic.RetagParams{
		From: body.From,
		To:   body.To,
	})
	if err != nil {
		h.WriteAPIError(w, err, logic.ErrRetagUnknownTag, logic.ErrRetagSameTag)

		return
	}

	h.WriteJSON(w, http.StatusOK, apiRetagResponse{
		Tag:      apiTag{ID: result.Tag.ID, Name: result.Tag.Name},
		Retagged: result.Retagged,
	})
}

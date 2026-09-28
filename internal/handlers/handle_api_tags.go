package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

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

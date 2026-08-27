package handlers

import "net/http"

// apiCategory is deliberately thin: categories are a shared lookup table
// (CLAUDE.md), and every resource form that references one only needs its id
// and name.
type apiCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type apiCategoryListResponse struct {
	Data []apiCategory `json:"data"`
}

// GetAPICategories lists every category. Unlike every other /api/* resource
// this is not user-scoped — categories are global — so there is no per-user
// filtering to do.
func (h *Handler) GetAPICategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.store.FindCategories(r.Context())
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	data := make([]apiCategory, 0, len(categories))
	for _, category := range categories {
		data = append(data, apiCategory{ID: category.ID, Name: category.Name})
	}

	h.WriteJSON(w, http.StatusOK, apiCategoryListResponse{Data: data})
}

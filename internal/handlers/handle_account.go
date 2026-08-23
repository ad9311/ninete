package handlers

import (
	"net/http"
)

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, http.StatusOK, AccountIndex)
}

package handlers

import (
	"net/http"
)

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, http.StatusOK, DashboardIndex)
}

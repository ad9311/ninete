package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/webtmpl"
)

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, webtmpl.DashboardIndex, h.tmplData(r))
}

package handlers

import "net/http"

// GetApp serves the SPA shell. It answers every path under /app/* with the
// same document (docs/spa-migration.md §3.7's client router owns the rest of
// the path), so a hard refresh on a nested route such as /app/expenses/12/edit
// resolves instead of 404ing.
func (h *Handler) GetApp(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, http.StatusOK, AppIndex)
}

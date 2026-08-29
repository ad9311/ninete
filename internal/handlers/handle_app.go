package handlers

import "net/http"

// GetApp serves the SPA shell. Since Phase 7 of docs/spa-migration.md it
// backs the catch-all at "/", answering every non-API, non-static path with
// the same document (§3.7's client router owns the rest of the path), so a
// hard refresh on a nested route such as /expenses/12/edit resolves instead
// of 404ing — and an unknown path gets the client router's own "Not found"
// rather than the server's.
func (h *Handler) GetApp(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, http.StatusOK, AppIndex)
}

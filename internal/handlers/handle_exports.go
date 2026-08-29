package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GetExportsExpenses streams the expense export (§5, Phase 5). It is reached
// through a plain anchor rather than lib/api.ts, since a fetch response cannot
// be handed to the browser's save flow the way a real navigation can.
//
// It sits on the *page* chain, not under /api, precisely because a plain
// anchor is a browser navigation: AuthMiddleware answers an expired session
// with a redirect to the login page, which is the only thing a navigation can
// act on. The API chain's 401 carries no Location, so the browser had nothing
// to follow and saved the JSON error envelope as expenses.json instead.
//
// The response is still JSON, so an error is still reported with the JSON
// writers rather than a rendered page — nothing here has an HTML document to
// fall back to.
func (h *Handler) GetExportsExpenses(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)

	expenses, err := h.store.ExportExpenses(r.Context(), user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	now := time.Now().UTC().Unix()
	payload := map[string]any{
		"exported_at": now,
		"expenses":    expenses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="expenses-%d.json"`, now))
	w.WriteHeader(http.StatusOK)

	// Streamed rather than buffered: the export is unpaginated, so buffering
	// it would hold every expense the account has in memory to compute a
	// Content-Length nothing needs.
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		h.app.Logger.Errorf("failed to write expenses export: %v", err)
	}
}

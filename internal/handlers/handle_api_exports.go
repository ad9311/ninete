package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GetAPIExportsExpenses streams the expense export (§5, Phase 5). It is
// reached through a plain anchor rather than lib/api.ts, since a fetch
// response cannot be handed to the browser's save flow the way a real
// navigation can.
func (h *Handler) GetAPIExportsExpenses(w http.ResponseWriter, r *http.Request) {
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

	// Streamed rather than buffered: see GetExportsExpenses for why.
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		h.app.Logger.Errorf("failed to write expenses export: %v", err)
	}
}

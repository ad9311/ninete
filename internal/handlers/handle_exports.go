package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (h *Handler) GetExports(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, http.StatusOK, ExportsIndex)
}

func (h *Handler) GetExportsExpenses(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)

	expenses, err := h.store.ExportExpenses(r.Context(), user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

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

	// Streamed rather than buffered: the payload is plain structs, so encoding
	// can only fail on the write itself, which a buffer would not have saved
	// either. Nothing is held in memory a second time.
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		h.app.Logger.Errorf("failed to write expenses export: %v", err)
	}
}

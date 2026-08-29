package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

type quickExpenseRequestBody struct {
	QuickInput string `json:"quick_input"`
	CategoryID int    `json:"category_id"`
	// TZOffset is Date.getTimezoneOffset() from the client, sent explicitly.
	// This is quick-add's own zone consumer (§3.6 of docs/spa-migration.md,
	// "Consumer 2") and was not retired alongside the named date ranges: it has
	// nothing to do with them, and dropping it would silently resolve
	// "today"/"yesterday" against UTC for any non-UTC user.
	TZOffset int `json:"tz_offset"`
}

// PostAPIExpensesQuick creates an expense from the quick-add input. When the
// description has no remembered category yet, it answers 422 with a
// "category_id" field entry rather than creating anything — the client shows a
// category picker and resubmits with category_id set.
func (h *Handler) PostAPIExpensesQuick(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	var body quickExpenseRequestBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	parsed, err := logic.ParseQuickExpense(body.QuickInput, body.TZOffset)
	if err != nil {
		h.WriteAPIError(w, err,
			logic.ErrQuickExpenseFormat,
			logic.ErrQuickExpenseDescription,
			logic.ErrQuickExpenseAmount,
			logic.ErrQuickExpenseDate,
			logic.ErrQuickExpenseTags,
			logic.ErrQuickExpenseTagName,
		)

		return
	}

	categoryID := body.CategoryID
	if categoryID == 0 {
		resolvedID, found, err := h.store.ResolveQuickExpenseCategory(ctx, user.ID, parsed.Description)
		if err != nil {
			h.WriteAPIError(w, err)

			return
		}
		if !found {
			h.WriteJSON(w, http.StatusUnprocessableEntity, APIError{
				Error:  "category required for this description",
				Fields: map[string]string{"category_id": "required"},
			})

			return
		}
		categoryID = resolvedID
	}

	expense, err := h.store.CreateQuickExpense(ctx, user.ID, categoryID, parsed)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.respondWithExpense(w, r, expense)
}

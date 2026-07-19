package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

// PostExpensesQuick handles the quick-add expense form: a single free-text field
// ("description, amount, date"). It resolves the category from a remembered
// mapping; on the first use of a description it re-renders the form asking the
// user to pick a category, which is then saved for future reuse.
func (h *Handler) PostExpensesQuick(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	if err := r.ParseForm(); err != nil {
		h.renderQuickErr(w, r, "", err)

		return
	}

	rawInput := r.FormValue("quick_input")
	categories, _, categoriesErr := h.findCategories(ctx)
	setExpenseFormData(data, categories, repo.Expense{}, "")
	setQuickFormData(data, categories, rawInput, false)

	parsed, err := logic.ParseQuickExpense(rawInput, parseTZOffset(r))
	if err != nil {
		h.renderQuickErr(w, r, rawInput, err)

		return
	}

	categoryID, err := parseOptionalCategoryID(r)
	if err != nil {
		h.renderQuickErr(w, r, rawInput, err)

		return
	}

	if categoryID == 0 {
		resolvedID, found, resolveErr := h.store.ResolveQuickExpenseCategory(ctx, user.ID, parsed.Description)
		if resolveErr != nil {
			h.renderQuickErr(w, r, rawInput, resolveErr)

			return
		}
		if !found {
			setQuickFormData(data, categories, rawInput, true)
			// Turbo only renders a form response on non-2xx; 422 signals "re-render".
			h.render(w, http.StatusUnprocessableEntity, ExpensesNew, data)

			return
		}
		categoryID = resolvedID
	}

	if _, err := h.store.CreateQuickExpense(ctx, user.ID, categoryID, parsed); err != nil {
		h.renderQuickErr(w, r, rawInput, err)

		return
	}

	if categoriesErr != nil {
		h.app.Logger.Errorf("failed to load categories: %v", categoriesErr)
	}

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// renderQuickErr re-renders the new-expense page with the quick form active,
// preserving the raw input and showing the error message.
func (h *Handler) renderQuickErr(w http.ResponseWriter, r *http.Request, rawInput string, err error) {
	data := h.tmplData(r)
	categories, _, _ := h.findCategories(r.Context())
	setExpenseFormData(data, categories, repo.Expense{}, "")
	setQuickFormData(data, categories, rawInput, false)
	h.renderErr(w, r, http.StatusBadRequest, ExpensesNew, err)
}

func parseOptionalCategoryID(r *http.Request) (int, error) {
	raw := r.FormValue("category_id")
	if raw == "" {
		return 0, nil
	}

	return prog.ParseID(raw, "Category ID")
}

func setQuickFormData(
	data map[string]any,
	categories []repo.Category,
	quickInput string,
	quickNeedsCategory bool,
) {
	data["categories"] = categories
	data["quickInput"] = quickInput
	data["quickNeedsCategory"] = quickNeedsCategory
	data["quickActive"] = quickInput != "" || quickNeedsCategory
}

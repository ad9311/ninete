package handlers

import (
	"fmt"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

func (h *Handler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	user := getCurrentUser(r)

	userIDFilter := repo.FilterField{
		Name:     "user_id",
		Value:    user.ID,
		Operator: "=",
	}

	opts := repo.QueryOptions{}
	opts.Filters.FilterFields = append(opts.Filters.FilterFields, userIDFilter)
	opts.Filters.Connector = "AND"

	expenses, err := h.store.FindExpenses(r.Context(), opts)
	if err != nil {
		data["error"] = err.Error()
	}

	data["expenses"] = expenses

	h.render(w, http.StatusOK, ExpensesIndex, data)
}

func (h *Handler) GetExpensesNew(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, ExpensesNew, h.tmplData(r))
}

func (h *Handler) PostExpenses(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)

	params, err := parseExpenseForm(r)
	if err != nil {
		data["error"] = err.Error()
		h.render(w, http.StatusBadRequest, ExpensesNew, h.tmplData(r))

		return
	}

	ctx := r.Context()
	user := getCurrentUser(r)

	_, err = h.store.CreateExpense(ctx, user.ID, params)
	if err != nil {
		data["error"] = err.Error()
		h.render(w, http.StatusBadRequest, ExpensesNew, data)

		return
	}

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func parseExpenseForm(r *http.Request) (logic.ExpenseParams, error) {
	var params logic.ExpenseParams
	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("failed to parse form, %w", err)
	}

	categoryID, err := prog.ParseID(r.FormValue("category_id"), "Category ID")
	if err != nil {
		return params, err
	}

	amount, err := prog.ParseAmount(r.FormValue("amount"))
	if err != nil {
		return params, err
	}

	date, err := prog.StringToUnixDate(r.FormValue("date"))
	if err != nil {
		return params, err
	}

	params.CategoryID = categoryID
	params.Description = r.FormValue("description")
	params.Amount = amount
	params.Date = date

	return params, nil
}

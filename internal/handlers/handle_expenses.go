package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

type expenseRow struct {
	ID           int
	CategoryName string
	Description  string
	Amount       uint64
	Date         int64
	Tags         string
}

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

func (h *Handler) ExpenseContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)
		expenseID := chi.URLParam(r, "id")

		id, err := prog.ParseID(expenseID, "Expense")
		if err != nil {
			h.NotFound(w, r)

			return
		}

		expense, err := h.store.FindExpense(ctx, id, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		if err != nil {
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

			return
		}

		ctx = context.WithValue(ctx, KeyExpense, &expense)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	user := getCurrentUser(r)

	q := r.URL.Query()
	sortOrder := q.Get("sort_order")
	sortField := q.Get("sort_field")

	userIDFilter := repo.FilterField{
		Name:     "user_id",
		Value:    user.ID,
		Operator: "=",
	}

	opts := repo.QueryOptions{
		Sorting: repo.Sorting{
			Order: sortOrder,
			Field: sortField,
		},
	}
	opts.Filters.FilterFields = append(opts.Filters.FilterFields, userIDFilter)
	opts.Filters.Connector = "AND"

	expenses, err := h.store.FindExpenses(r.Context(), opts)
	if err != nil {
		h.renderErr(w, r, http.StatusBadRequest, ExpensesIndex, err)

		return
	}

	_, categoryNameByID, err := h.findCategories(r.Context())
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ExpensesIndex, err)

		return
	}

	rows := make([]expenseRow, 0, len(expenses))
	expenseIDs := make([]int, 0, len(expenses))
	for _, expense := range expenses {
		expenseIDs = append(expenseIDs, expense.ID)
	}

	expenseTagRows, err := h.store.FindExpenseTagRows(r.Context(), expenseIDs, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ExpensesIndex, err)

		return
	}
	expenseTagNames := map[int][]string{}
	for _, row := range expenseTagRows {
		expenseTagNames[row.ExpenseID] = append(expenseTagNames[row.ExpenseID], row.TagName)
	}

	for _, expense := range expenses {
		categoryName := categoryNameByID[expense.CategoryID]
		if categoryName == "" {
			categoryName = "Unknown"
		}

		rows = append(rows, expenseRow{
			ID:           expense.ID,
			CategoryName: categoryName,
			Description:  expense.Description,
			Amount:       expense.Amount,
			Date:         expense.Date,
			Tags:         logic.JoinTagNames(expenseTagNames[expense.ID]),
		})
	}

	data["expenses"] = rows

	h.render(w, http.StatusOK, ExpensesIndex, data)
}

func (h *Handler) GetExpensesNew(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)

	categories, _, err := h.findCategories(r.Context())
	setExpenseFormData(data, categories, repo.Expense{}, "")
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ExpensesNew, err)

		return
	}

	h.render(w, http.StatusOK, ExpensesNew, data)
}

func (h *Handler) GetExpensesEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	expense := getExpense(r)

	categories, _, err := h.findCategories(ctx)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ExpensesEdit, err)

		return
	}

	expenseTags, err := h.store.FindExpenseTags(ctx, expense.ID, getCurrentUser(r).ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ExpensesEdit, err)

		return
	}

	var tagNames []string
	for _, tag := range expenseTags {
		tagNames = append(tagNames, tag.Name)
	}
	setExpenseFormData(data, categories, *expense, logic.JoinTagNames(tagNames))

	h.render(w, http.StatusOK, ExpensesEdit, data)
}

func (h *Handler) PostExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	rawTagsInput := r.FormValue("tags")

	categories, _, categoriesErr := h.findCategories(ctx)
	setExpenseFormData(data, categories, repo.Expense{}, rawTagsInput)

	params, err := parseExpenseForm(r)
	if err != nil {
		h.renderErr(w, r, http.StatusBadRequest, ExpensesNew, err)

		return
	}

	user := getCurrentUser(r)

	_, err = h.store.CreateExpense(ctx, user.ID, params)
	if err != nil {
		setExpenseFormData(data, categories, repo.Expense{
			CategoryID:  params.CategoryID,
			Description: params.Description,
			Amount:      params.Amount,
			Date:        params.Date,
		}, logic.JoinTagNames(params.Tags))
		h.renderErr(w, r, http.StatusBadRequest, ExpensesNew, err)

		return
	}

	if categoriesErr != nil {
		h.app.Logger.Errorf("failed to load categories: %v", categoriesErr)
	}

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *Handler) PostExpensesUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	expense := *getExpense(r)
	rawTagsInput := r.FormValue("tags")

	categories, _, categoriesErr := h.findCategories(ctx)
	setExpenseFormData(data, categories, expense, rawTagsInput)

	params, err := parseExpenseForm(r)
	if err != nil {
		h.renderErr(w, r, http.StatusBadRequest, ExpensesEdit, err)

		return
	}

	_, err = h.store.UpdateExpense(ctx, expense.ID, user.ID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		expense.CategoryID = params.CategoryID
		expense.Description = params.Description
		expense.Amount = params.Amount
		expense.Date = params.Date
		setExpenseFormData(data, categories, expense, logic.JoinTagNames(params.Tags))
		h.renderErr(w, r, http.StatusBadRequest, ExpensesEdit, err)

		return
	}

	if categoriesErr != nil {
		h.app.Logger.Errorf("failed to load categories: %v", categoriesErr)
	}

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *Handler) PostExpensesDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	expense := getExpense(r)

	_, err := h.store.DeleteExpense(ctx, expense.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

func parseExpenseForm(r *http.Request) (logic.ExpenseParams, error) {
	var params logic.ExpenseParams
	base, err := parseExpenseFormBase(r)
	if err != nil {
		return params, err
	}

	date, err := prog.StringToUnixDate(r.FormValue("date"))
	if err != nil {
		return params, err
	}

	params.CategoryID = base.CategoryID
	params.Description = base.Description
	params.Amount = base.Amount
	params.Date = date
	params.Tags = logic.ParseTagNames(r.FormValue("tags"))

	return params, nil
}

func setExpenseFormData(
	data map[string]any,
	categories []repo.Category,
	expense repo.Expense,
	tagsInput string,
) {
	setResourceFormData(data, categories, "expense", expense)
	data["tagsInput"] = tagsInput
}

func getExpense(r *http.Request) *repo.Expense {
	expense, ok := r.Context().Value(KeyExpense).(*repo.Expense)

	if !ok {
		panic("failed to get expense context")
	}

	return expense
}

package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

type recurrentExpenseRow struct {
	ID           int
	CategoryName string
	Description  string
	Amount       uint64
	Period       uint
}

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

func (h *Handler) RecurrentExpenseContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)
		recurrentExpenseID := chi.URLParam(r, "id")

		id, err := prog.ParseID(recurrentExpenseID, "Recurrent expense")
		if err != nil {
			h.NotFound(w, r)

			return
		}

		recurrentExpense, err := h.store.FindRecurrentExpense(ctx, id, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		if err != nil {
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

			return
		}

		ctx = context.WithValue(ctx, KeyRecurrentExpense, &recurrentExpense)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
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

	recurrentExpenses, err := h.store.FindRecurrentExpenses(r.Context(), opts)
	if err != nil {
		h.renderErr(w, r, http.StatusBadRequest, RecurrentExpensesIndex, err)

		return
	}

	_, categoryNameByID, err := h.findCategories(r.Context())
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, RecurrentExpensesIndex, err)

		return
	}

	rows := make([]recurrentExpenseRow, 0, len(recurrentExpenses))
	for _, recurrentExpense := range recurrentExpenses {
		categoryName := categoryNameByID[recurrentExpense.CategoryID]
		if categoryName == "" {
			categoryName = "Unknown"
		}

		rows = append(rows, recurrentExpenseRow{
			ID:           recurrentExpense.ID,
			CategoryName: categoryName,
			Description:  recurrentExpense.Description,
			Amount:       recurrentExpense.Amount,
			Period:       recurrentExpense.Period,
		})
	}

	data["recurrentExpenses"] = rows

	h.render(w, http.StatusOK, RecurrentExpensesIndex, data)
}

func (h *Handler) GetRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	recurrentExpense := getRecurrentExpense(r)

	_, categoryNameByID, err := h.findCategories(ctx)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, RecurrentExpensesShow, err)

		return
	}

	categoryName := categoryNameByID[recurrentExpense.CategoryID]
	if categoryName == "" {
		categoryName = "Unknown"
	}

	data["recurrentExpense"] = recurrentExpenseRow{
		ID:           recurrentExpense.ID,
		CategoryName: categoryName,
		Description:  recurrentExpense.Description,
		Amount:       recurrentExpense.Amount,
		Period:       recurrentExpense.Period,
	}

	h.render(w, http.StatusOK, RecurrentExpensesShow, data)
}

func (h *Handler) GetRecurrentExpensesNew(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)

	categories, _, err := h.findCategories(r.Context())
	setRecurrentExpenseFormData(data, categories, repo.RecurrentExpense{})
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, RecurrentExpensesNew, err)

		return
	}

	h.render(w, http.StatusOK, RecurrentExpensesNew, data)
}

func (h *Handler) GetRecurrentExpensesEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	recurrentExpense := getRecurrentExpense(r)

	categories, _, err := h.findCategories(ctx)
	setRecurrentExpenseFormData(data, categories, *recurrentExpense)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, RecurrentExpensesEdit, err)

		return
	}

	h.render(w, http.StatusOK, RecurrentExpensesEdit, data)
}

func (h *Handler) PostRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)

	categories, _, categoriesErr := h.findCategories(ctx)
	setRecurrentExpenseFormData(data, categories, repo.RecurrentExpense{})

	params, err := parseRecurrentExpenseForm(r)
	if err != nil {
		h.renderErr(w, r, http.StatusBadRequest, RecurrentExpensesNew, err)

		return
	}

	user := getCurrentUser(r)

	_, err = h.store.CreateRecurrentExpense(ctx, user.ID, params)
	if err != nil {
		setRecurrentExpenseFormData(data, categories, repo.RecurrentExpense{
			CategoryID:  params.CategoryID,
			Description: params.Description,
			Amount:      params.Amount,
			Period:      params.Period,
		})
		h.renderErr(w, r, http.StatusBadRequest, RecurrentExpensesNew, err)

		return
	}

	if categoriesErr != nil {
		h.app.Logger.Errorf("failed to load categories: %v", categoriesErr)
	}

	http.Redirect(w, r, "/recurrent-expenses", http.StatusSeeOther)
}

func (h *Handler) PostRecurrentExpensesUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	recurrentExpense := *getRecurrentExpense(r)

	categories, _, categoriesErr := h.findCategories(ctx)
	setRecurrentExpenseFormData(data, categories, recurrentExpense)

	params, err := parseRecurrentExpenseForm(r)
	if err != nil {
		h.renderErr(w, r, http.StatusBadRequest, RecurrentExpensesEdit, err)

		return
	}

	_, err = h.store.UpdateRecurrentExpense(ctx, recurrentExpense.ID, user.ID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		recurrentExpense.CategoryID = params.CategoryID
		recurrentExpense.Description = params.Description
		recurrentExpense.Amount = params.Amount
		recurrentExpense.Period = params.Period
		setRecurrentExpenseFormData(data, categories, recurrentExpense)
		h.renderErr(w, r, http.StatusBadRequest, RecurrentExpensesEdit, err)

		return
	}

	if categoriesErr != nil {
		h.app.Logger.Errorf("failed to load categories: %v", categoriesErr)
	}

	http.Redirect(w, r, "/recurrent-expenses", http.StatusSeeOther)
}

func (h *Handler) PostRecurrentExpensesDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	recurrentExpense := getRecurrentExpense(r)

	_, err := h.store.DeleteRecurrentExpense(ctx, recurrentExpense.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/recurrent-expenses", http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

func parseRecurrentExpenseForm(r *http.Request) (logic.RecurrentExpenseParams, error) {
	var params logic.RecurrentExpenseParams

	base, err := parseExpenseFormBase(r)
	if err != nil {
		return params, err
	}

	period, err := prog.ParseID(r.FormValue("period"), "Period")
	if err != nil {
		return params, err
	}
	if period < 1 {
		return params, fmt.Errorf("%w of Period \"%v\", period cannot be lower than 1", prog.ErrParsing, period)
	}

	params.CategoryID = base.CategoryID
	params.Description = base.Description
	params.Amount = base.Amount
	params.Period = uint(period)

	return params, nil
}

func setRecurrentExpenseFormData(
	data map[string]any,
	categories []repo.Category,
	recurrentExpense repo.RecurrentExpense,
) {
	setResourceFormData(data, categories, "recurrentExpense", recurrentExpense)
}

func getRecurrentExpense(r *http.Request) *repo.RecurrentExpense {
	recurrentExpense, ok := r.Context().Value(KeyRecurrentExpense).(*repo.RecurrentExpense)

	if !ok {
		panic("failed to get recurrent expense context")
	}

	return recurrentExpense
}

package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

type recurrentExpenseRow struct {
	ID              int
	CategoryName    string
	Description     string
	Amount          uint64
	Period          uint
	OccurrenceLimit uint
	OccurrenceCount uint
	Archived        bool
	Tags            []string
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
	h.renderRecurrentExpenseList(w, r, false, RecurrentExpensesIndex, "/recurrent-expenses")
}

// GetRecurrentExpensesArchived lists the rows the cron job stopped copying
// because they hit their occurrence limit. They stay out of the main list until
// the owner unarchives one.
func (h *Handler) GetRecurrentExpensesArchived(w http.ResponseWriter, r *http.Request) {
	h.renderRecurrentExpenseList(w, r, true, RecurrentExpensesArchived, "/recurrent-expenses/archived")
}

func (h *Handler) renderRecurrentExpenseList(
	w http.ResponseWriter,
	r *http.Request,
	archived bool,
	template TemplateName,
	basePath string,
) {
	data := h.tmplData(r)
	user := getCurrentUser(r)

	opts := userScopedQueryOpts(r, user.ID, repo.Sorting{Field: "created_at", Order: "DESC"}, "")
	opts.Filters.FilterFields = append(opts.Filters.FilterFields, repo.RecurrentExpenseArchivedFilter(archived))

	totalCount, err := h.store.CountRecurrentExpenses(r.Context(), opts.Filters)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, template, err)

		return
	}

	recurrentExpenses, err := h.store.FindRecurrentExpenses(r.Context(), opts)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, template, err)

		return
	}

	categories, categoryNameByID, ok := h.findCategoriesOrErr(w, r, template)
	if !ok {
		return
	}

	recurrentExpenseIDs := make([]int, 0, len(recurrentExpenses))
	for _, recurrentExpense := range recurrentExpenses {
		recurrentExpenseIDs = append(recurrentExpenseIDs, recurrentExpense.ID)
	}

	tagRows, err := h.store.FindTagRows(
		r.Context(),
		repo.TaggableTypeRecurrentExpense,
		"recurrent_expenses",
		recurrentExpenseIDs,
		user.ID,
	)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, template, err)

		return
	}
	tagNames := repo.TagNamesByTargetID(tagRows)

	rows := make([]recurrentExpenseRow, 0, len(recurrentExpenses))
	for _, recurrentExpense := range recurrentExpenses {
		rows = append(rows, newRecurrentExpenseRow(
			recurrentExpense,
			categoryNameOrUnknown(categoryNameByID, recurrentExpense.CategoryID),
			tagNames[recurrentExpense.ID],
		))
	}

	data["recurrentExpenses"] = rows
	data["categories"] = categories
	data["pagination"] = newPaginationData(r, opts, totalCount, "")
	data["basePath"] = basePath

	h.render(w, http.StatusOK, template, data)
}

func (h *Handler) GetRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	recurrentExpense := getRecurrentExpense(r)

	_, categoryNameByID, ok := h.findCategoriesOrErr(w, r, RecurrentExpensesShow)
	if !ok {
		return
	}

	tags, err := h.store.FindRecurrentExpenseTags(r.Context(), recurrentExpense.ID, getCurrentUser(r).ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, RecurrentExpensesShow, err)

		return
	}

	data["recurrentExpense"] = newRecurrentExpenseRow(
		*recurrentExpense,
		categoryNameOrUnknown(categoryNameByID, recurrentExpense.CategoryID),
		logic.ExtractTagNames(tags),
	)

	h.render(w, http.StatusOK, RecurrentExpensesShow, data)
}

func (h *Handler) GetRecurrentExpensesNew(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)

	categories, _, ok := h.findCategoriesOrErr(w, r, RecurrentExpensesNew)
	if !ok {
		return
	}

	setRecurrentExpenseFormData(data, categories, repo.RecurrentExpense{}, "")

	h.render(w, http.StatusOK, RecurrentExpensesNew, data)
}

func (h *Handler) GetRecurrentExpensesEdit(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	recurrentExpense := getRecurrentExpense(r)

	categories, _, ok := h.findCategoriesOrErr(w, r, RecurrentExpensesEdit)
	if !ok {
		return
	}

	tags, err := h.store.FindRecurrentExpenseTags(r.Context(), recurrentExpense.ID, getCurrentUser(r).ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, RecurrentExpensesEdit, err)

		return
	}

	setRecurrentExpenseFormData(
		data,
		categories,
		*recurrentExpense,
		logic.JoinTagNames(logic.ExtractTagNames(tags)),
	)

	h.render(w, http.StatusOK, RecurrentExpensesEdit, data)
}

func (h *Handler) PostRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)

	rawTagsInput := r.FormValue("tags")

	categories, _, categoriesErr := h.findCategories(ctx)
	setRecurrentExpenseFormData(data, categories, repo.RecurrentExpense{}, rawTagsInput)

	params, err := parseRecurrentExpenseForm(r)
	if err != nil {
		h.renderErr(w, r, http.StatusBadRequest, RecurrentExpensesNew, err)

		return
	}

	user := getCurrentUser(r)

	_, err = h.store.CreateRecurrentExpense(ctx, user.ID, params)
	if err != nil {
		setRecurrentExpenseFormData(data, categories, repo.RecurrentExpense{
			CategoryID:      params.CategoryID,
			Description:     params.Description,
			Amount:          params.Amount,
			Period:          params.Period,
			OccurrenceLimit: params.OccurrenceLimit,
		}, logic.JoinTagNames(params.Tags))
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

	rawTagsInput := r.FormValue("tags")

	categories, _, categoriesErr := h.findCategories(ctx)
	setRecurrentExpenseFormData(data, categories, recurrentExpense, rawTagsInput)

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
		recurrentExpense.OccurrenceLimit = params.OccurrenceLimit
		setRecurrentExpenseFormData(data, categories, recurrentExpense, logic.JoinTagNames(params.Tags))
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

// PostRecurrentExpensesUnarchive puts an archived recurrent expense back in
// rotation. Editing one never does this on its own — the owner has to ask.
func (h *Handler) PostRecurrentExpensesUnarchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	recurrentExpense := getRecurrentExpense(r)

	_, err := h.store.UnarchiveRecurrentExpense(ctx, recurrentExpense.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/recurrent-expenses/%d", recurrentExpense.ID), http.StatusSeeOther)
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

	occurrenceLimit, err := parseOccurrenceLimit(r.FormValue("occurrence_limit"))
	if err != nil {
		return params, err
	}

	params.CategoryID = base.CategoryID
	params.Description = base.Description
	params.Amount = base.Amount
	params.Period = uint(period)
	params.OccurrenceLimit = occurrenceLimit
	params.Tags = logic.ParseTagNames(r.FormValue("tags"))

	return params, nil
}

// parseOccurrenceLimit accepts an empty field as "unlimited" so an existing form
// submitted without the input keeps the old behaviour.
func parseOccurrenceLimit(value string) (uint, error) {
	if value == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%w of Occurrence limit \"%v\"", prog.ErrParsing, value)
	}

	if limit < 0 {
		return 0, fmt.Errorf(
			"%w of Occurrence limit \"%v\", occurrence limit cannot be negative",
			prog.ErrParsing,
			limit,
		)
	}

	return uint(limit), nil
}

func newRecurrentExpenseRow(
	recurrentExpense repo.RecurrentExpense,
	categoryName string,
	tags []string,
) recurrentExpenseRow {
	return recurrentExpenseRow{
		ID:              recurrentExpense.ID,
		CategoryName:    categoryName,
		Description:     recurrentExpense.Description,
		Amount:          recurrentExpense.Amount,
		Period:          recurrentExpense.Period,
		OccurrenceLimit: recurrentExpense.OccurrenceLimit,
		OccurrenceCount: recurrentExpense.OccurrenceCount,
		Archived:        recurrentExpense.ArchivedAt != nil,
		Tags:            tags,
	}
}

func setRecurrentExpenseFormData(
	data map[string]any,
	categories []repo.Category,
	recurrentExpense repo.RecurrentExpense,
	tagsInput string,
) {
	setResourceFormData(data, categories, "recurrentExpense", recurrentExpense)
	data["tagsInput"] = tagsInput
}

func getRecurrentExpense(r *http.Request) *repo.RecurrentExpense {
	recurrentExpense, ok := r.Context().Value(KeyRecurrentExpense).(*repo.RecurrentExpense)

	if !ok {
		panic("failed to get recurrent expense context")
	}

	return recurrentExpense
}

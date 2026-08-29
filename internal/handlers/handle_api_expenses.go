package handlers

import (
	"context"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

// apiExpense is the expense as the client sees it — a separate type from
// repo.Expense so a column added there is not published by accident (§3.5 of
// docs/spa-migration.md). Date is a calendar date (UTC midnight), CreatedAt an
// instant — see §3.6 before touching either.
type apiExpense struct {
	ID           int      `json:"id"`
	CategoryID   int      `json:"category_id"`
	CategoryName string   `json:"category_name"`
	Description  string   `json:"description"`
	Amount       uint64   `json:"amount"`
	Date         int64    `json:"date"`
	CreatedAt    int64    `json:"created_at"`
	Tags         []string `json:"tags"`
}

func newAPIExpense(expense repo.Expense, categoryName string, tags []string) apiExpense {
	return apiExpense{
		ID:           expense.ID,
		CategoryID:   expense.CategoryID,
		CategoryName: categoryName,
		Description:  expense.Description,
		Amount:       expense.Amount,
		Date:         expense.Date,
		CreatedAt:    expense.CreatedAt,
		Tags:         emptyIfNil(tags),
	}
}

type apiExpenseListResponse struct {
	Data       []apiExpense  `json:"data"`
	Pagination apiPagination `json:"pagination"`
}

type expenseRequestBody struct {
	CategoryID  int      `json:"category_id"`
	Description string   `json:"description"`
	Amount      uint64   `json:"amount"`
	Date        int64    `json:"date"`
	Tags        []string `json:"tags"`
}

func (b expenseRequestBody) toParams() logic.ExpenseParams {
	return logic.ExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  b.CategoryID,
			Description: b.Description,
			Amount:      b.Amount,
		},
		Date: b.Date,
		Tags: logic.NormalizeTagNames(b.Tags),
	}
}

// apiExpenseListOpts builds the /api/expenses query options: it delegates
// sorting, pagination and the category filter to userScopedQueryOpts, which
// applies no date filter of its own, and then layers on the explicit
// [start, end) bounds the client resolved. expenseSearch.apply layers its own
// predicates on top.
//
// It reports whether the request carried bounds, which the caller has to feed
// back into the search — see GetAPIExpenses.
func apiExpenseListOpts(r *http.Request, userID int) (repo.QueryOptions, bool, error) {
	opts := userScopedQueryOpts(r, userID, repo.Sorting{Field: "date", Order: "DESC"})

	start, end, hasBounds, err := parseAPIDateBounds(r.URL.Query())
	if err != nil {
		return opts, false, err
	}
	if hasBounds {
		opts.Filters.FilterFields = append(opts.Filters.FilterFields,
			repo.FilterField{Name: "date", Value: start, Operator: ">="},
			repo.FilterField{Name: "date", Value: end, Operator: "<"},
		)
	}

	return opts, hasBounds, nil
}

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

// APIExpenseContext looks the expense up once for the /api/expenses/{id}
// subtree, storing it under KeyExpense for getExpense. A missing or malformed
// id answers the same 404 envelope every other /api/* failure uses.
func (h *Handler) APIExpenseContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)

		id, err := prog.ParseID(chi.URLParam(r, "id"), "Expense")
		if err != nil {
			h.APINotFound(w, r)

			return
		}

		expense, err := h.store.FindExpense(ctx, id, user.ID)
		if err != nil {
			h.WriteAPIError(w, err)

			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, KeyExpense, &expense)))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetAPIExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	search, err := parseExpenseSearch(r)
	if err != nil {
		h.WriteAPIError(w, err, ErrSearchTermTooLong, ErrSearchDateFormat, ErrSearchDateRange)

		return
	}

	opts, hasBounds, err := apiExpenseListOpts(r, user.ID)
	if err != nil {
		h.WriteAPIError(w, err, ErrAPIInvalidDateRange)

		return
	}

	// parseExpenseSearch reads explicitRange from date_range, which this chain
	// never receives: the client resolves its named range to bounds itself
	// (§3.6 of docs/spa-migration.md). Bounds present *is* the explicit range
	// here, and without this a text search would take clearsPresetRange's
	// implicit-widening branch and delete the very bounds the client asked for.
	search.explicitRange = hasBounds
	search.apply(&opts, user.ID)

	totalCount, err := h.store.CountExpenses(ctx, opts.Filters)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	expenses, err := h.store.FindExpenses(ctx, opts)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	_, categoryNameByID, err := h.findCategories(ctx)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	ids := make([]int, 0, len(expenses))
	for _, expense := range expenses {
		ids = append(ids, expense.ID)
	}

	tagRows, err := h.store.FindTagRows(ctx, repo.TaggableTypeExpense, "expenses", ids, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}
	tagNames := repo.TagNamesByTargetID(tagRows)

	data := make([]apiExpense, 0, len(expenses))
	for _, expense := range expenses {
		data = append(
			data,
			newAPIExpense(expense, categoryNameOrUnknown(categoryNameByID, expense.CategoryID), tagNames[expense.ID]),
		)
	}

	h.WriteJSON(w, http.StatusOK, apiExpenseListResponse{
		Data:       data,
		Pagination: newAPIPagination(newPaginationData(r, opts, totalCount)),
	})
}

func (h *Handler) GetAPIExpense(w http.ResponseWriter, r *http.Request) {
	h.respondWithExpense(w, r, *getExpense(r))
}

func (h *Handler) PostAPIExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	var body expenseRequestBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	expense, err := h.store.CreateExpense(ctx, user.ID, body.toParams())
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.respondWithExpense(w, r, expense)
}

// PutAPIExpense replaces the record — see PutAPIRecurrentExpense's comment for
// why this is PUT rather than PATCH.
func (h *Handler) PutAPIExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	expense := getExpense(r)

	var body expenseRequestBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	updated, err := h.store.UpdateExpense(ctx, expense.ID, user.ID, body.toParams())
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.respondWithExpense(w, r, updated)
}

// respondWithExpense re-reads the tags rather than echoing the request body,
// same reasoning as respondWithRecurrentExpense: normalization can change what
// was sent.
func (h *Handler) respondWithExpense(w http.ResponseWriter, r *http.Request, expense repo.Expense) {
	ctx := r.Context()
	user := getCurrentUser(r)

	_, categoryNameByID, err := h.findCategories(ctx)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	tags, err := h.store.FindExpenseTags(ctx, expense.ID, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.WriteJSON(w, http.StatusOK, newAPIExpense(
		expense,
		categoryNameOrUnknown(categoryNameByID, expense.CategoryID),
		logic.ExtractTagNames(tags),
	))
}

func (h *Handler) DeleteAPIExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	expense := getExpense(r)

	if _, err := h.store.DeleteExpense(ctx, expense.ID, user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

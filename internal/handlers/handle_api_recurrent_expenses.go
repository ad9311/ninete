package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

// apiRecurrentExpense is the recurrent expense as the client sees it — a
// separate type from repo.RecurrentExpense so a column added there is not
// published by accident (§3.5 of docs/spa-migration.md).
type apiRecurrentExpense struct {
	ID              int      `json:"id"`
	CategoryID      int      `json:"category_id"`
	CategoryName    string   `json:"category_name"`
	Description     string   `json:"description"`
	Amount          uint64   `json:"amount"`
	Period          uint     `json:"period"`
	OccurrenceLimit uint     `json:"occurrence_limit"`
	OccurrenceCount uint     `json:"occurrence_count"`
	Archived        bool     `json:"archived"`
	Tags            []string `json:"tags"`
}

func newAPIRecurrentExpense(re repo.RecurrentExpense, categoryName string, tags []string) apiRecurrentExpense {
	return apiRecurrentExpense{
		ID:              re.ID,
		CategoryID:      re.CategoryID,
		CategoryName:    categoryName,
		Description:     re.Description,
		Amount:          re.Amount,
		Period:          re.Period,
		OccurrenceLimit: re.OccurrenceLimit,
		OccurrenceCount: re.OccurrenceCount,
		Archived:        re.ArchivedAt != nil,
		Tags:            emptyIfNil(tags),
	}
}

// emptyIfNil keeps a tagless resource's "tags" field an empty JSON array
// rather than null — the client always expects something it can map over.
func emptyIfNil(tags []string) []string {
	if tags == nil {
		return []string{}
	}

	return tags
}

type apiPagination struct {
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	PerPage     int    `json:"per_page"`
	TotalCount  int    `json:"total_count"`
	HasPrev     bool   `json:"has_prev"`
	HasNext     bool   `json:"has_next"`
	SortField   string `json:"sort_field"`
	SortOrder   string `json:"sort_order"`
	CategoryID  int    `json:"category_id"`
}

func newAPIPagination(p PaginationData) apiPagination {
	return apiPagination{
		CurrentPage: p.CurrentPage,
		TotalPages:  p.TotalPages,
		PerPage:     p.PerPage,
		TotalCount:  p.TotalCount,
		HasPrev:     p.HasPrev,
		HasNext:     p.HasNext,
		SortField:   p.SortField,
		SortOrder:   p.SortOrder,
		CategoryID:  p.CategoryID,
	}
}

type apiRecurrentExpenseListResponse struct {
	Data       []apiRecurrentExpense `json:"data"`
	Pagination apiPagination         `json:"pagination"`
}

type recurrentExpenseRequestBody struct {
	CategoryID      int      `json:"category_id"`
	Description     string   `json:"description"`
	Amount          uint64   `json:"amount"`
	Period          uint     `json:"period"`
	OccurrenceLimit uint     `json:"occurrence_limit"`
	Tags            []string `json:"tags"`
}

func (b recurrentExpenseRequestBody) toParams() logic.RecurrentExpenseParams {
	return logic.RecurrentExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  b.CategoryID,
			Description: b.Description,
			Amount:      b.Amount,
		},
		Period:          b.Period,
		OccurrenceLimit: b.OccurrenceLimit,
		// Reuses the same normalization the form path gets from
		// logic.ParseTagNames (lowercase, trim, dedupe, drop empty) rather
		// than a second implementation of it here. The slice overload is the
		// one to call: joining on ";" and re-splitting would tear a tag that
		// contains a semicolon into two.
		Tags: logic.NormalizeTagNames(b.Tags),
	}
}

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

// APIRecurrentExpenseContext is RecurrentExpenseContext's JSON-answering
// twin: a malformed or missing id answers the same 404 envelope every other
// /api/* failure uses instead of the HTML not-found page.
func (h *Handler) APIRecurrentExpenseContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)

		id, err := prog.ParseID(chi.URLParam(r, "id"), "Recurrent expense")
		if err != nil {
			h.APINotFound(w, r)

			return
		}

		recurrentExpense, err := h.store.FindRecurrentExpense(ctx, id, user.ID)
		if err != nil {
			h.WriteAPIError(w, err)

			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, KeyRecurrentExpense, &recurrentExpense)))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

// GetAPIRecurrentExpenses lists a user's recurrent expenses. ?archived=true
// answers what /recurrent-expenses/archived renders on the template side —
// one endpoint rather than two routes, now that the client router owns the
// query string (§2.3 of docs/spa-migration.md).
func (h *Handler) GetAPIRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	archived, _ := strconv.ParseBool(r.URL.Query().Get("archived"))

	opts := userScopedQueryOpts(r, user.ID, repo.Sorting{Field: "created_at", Order: "DESC"}, "")
	opts.Filters.FilterFields = append(opts.Filters.FilterFields, repo.RecurrentExpenseArchivedFilter(archived))

	totalCount, err := h.store.CountRecurrentExpenses(ctx, opts.Filters)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	recurrentExpenses, err := h.store.FindRecurrentExpenses(ctx, opts)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	_, categoryNameByID, err := h.findCategories(ctx)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	ids := make([]int, 0, len(recurrentExpenses))
	for _, re := range recurrentExpenses {
		ids = append(ids, re.ID)
	}

	tagRows, err := h.store.FindTagRows(ctx, repo.TaggableTypeRecurrentExpense, "recurrent_expenses", ids, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}
	tagNames := repo.TagNamesByTargetID(tagRows)

	data := make([]apiRecurrentExpense, 0, len(recurrentExpenses))
	for _, re := range recurrentExpenses {
		data = append(
			data,
			newAPIRecurrentExpense(re, categoryNameOrUnknown(categoryNameByID, re.CategoryID), tagNames[re.ID]),
		)
	}

	h.WriteJSON(w, http.StatusOK, apiRecurrentExpenseListResponse{
		Data:       data,
		Pagination: newAPIPagination(newPaginationData(r, opts, totalCount, "")),
	})
}

func (h *Handler) GetAPIRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	recurrentExpense := getRecurrentExpense(r)

	_, categoryNameByID, err := h.findCategories(ctx)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	tags, err := h.store.FindRecurrentExpenseTags(ctx, recurrentExpense.ID, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.WriteJSON(w, http.StatusOK, newAPIRecurrentExpense(
		*recurrentExpense,
		categoryNameOrUnknown(categoryNameByID, recurrentExpense.CategoryID),
		logic.ExtractTagNames(tags),
	))
}

func (h *Handler) PostAPIRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	var body recurrentExpenseRequestBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	recurrentExpense, err := h.store.CreateRecurrentExpense(ctx, user.ID, body.toParams())
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.respondWithRecurrentExpense(w, r, recurrentExpense)
}

// PutAPIRecurrentExpense replaces the record. It is PUT rather than PATCH
// because recurrentExpenseRequestBody has no pointer fields: an omitted key
// decodes to the zero value and is written as one, so a body carrying only
// "amount" would clear the tags and reset occurrence_limit to 0 (unlimited).
// That matches the template form's POST, which also submits every field. A
// real partial update means pointer fields here and a merge in toParams.
func (h *Handler) PutAPIRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	recurrentExpense := getRecurrentExpense(r)

	var body recurrentExpenseRequestBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	updated, err := h.store.UpdateRecurrentExpense(ctx, recurrentExpense.ID, user.ID, body.toParams())
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.respondWithRecurrentExpense(w, r, updated)
}

// respondWithRecurrentExpense re-reads the tags rather than echoing the
// request body: logic.ParseTagNames normalizes (lowercase, trim, dedupe), so
// the response must carry what was actually stored, not what was sent.
func (h *Handler) respondWithRecurrentExpense(
	w http.ResponseWriter,
	r *http.Request,
	recurrentExpense repo.RecurrentExpense,
) {
	ctx := r.Context()
	user := getCurrentUser(r)

	_, categoryNameByID, err := h.findCategories(ctx)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	tags, err := h.store.FindRecurrentExpenseTags(ctx, recurrentExpense.ID, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.WriteJSON(w, http.StatusOK, newAPIRecurrentExpense(
		recurrentExpense,
		categoryNameOrUnknown(categoryNameByID, recurrentExpense.CategoryID),
		logic.ExtractTagNames(tags),
	))
}

func (h *Handler) DeleteAPIRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	recurrentExpense := getRecurrentExpense(r)

	if _, err := h.store.DeleteRecurrentExpense(ctx, recurrentExpense.ID, user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostAPIRecurrentExpenseUnarchive is the JSON twin of
// PostRecurrentExpensesUnarchive: puts an archived recurrent expense back in
// rotation. Editing one never does this on its own — the owner has to ask.
func (h *Handler) PostAPIRecurrentExpenseUnarchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	recurrentExpense := getRecurrentExpense(r)

	updated, err := h.store.UnarchiveRecurrentExpense(ctx, recurrentExpense.ID, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.respondWithRecurrentExpense(w, r, updated)
}

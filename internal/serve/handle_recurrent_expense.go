package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

type recurrentExpenseUpdatePayload struct {
	Description string `json:"description"`
	Amount      uint64 `json:"amount"`
	Period      uint   `json:"period"`
}

func (s *Server) ContextRecurrentExpense(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserContext(r)

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			s.respondError(w, http.StatusNotFound, ErrInvalidId)

			return
		}

		recurrent, err := s.store.FindRecurrentExpense(r.Context(), id, user.ID)
		if err != nil {
			if errors.Is(err, logic.ErrNotFound) {
				s.respondError(w, http.StatusNotFound, err)

				return
			}
			s.respondError(w, http.StatusBadRequest, err)

			return
		}

		ctx := context.WithValue(r.Context(), prog.KeyRecurrentExpense, &recurrent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) GetRecurrentExpenses(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query().Get(queryOptionsKey)
	opts, err := decodeQueryOptions(params)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	user := getUserContext(r)
	userIDFilter := repo.FilterField{
		Name:     "user_id",
		Value:    user.ID,
		Operator: "=",
	}
	opts.Filters.FilterFields = append(opts.Filters.FilterFields, userIDFilter)
	opts.Filters.Connector = "AND"

	recurrentExpenses, err := s.store.FindRecurrentExpenses(r.Context(), opts)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	count, err := s.store.CountRecurrentExpenses(r.Context(), opts.Filters)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respondMeta(w, http.StatusOK, recurrentExpenses, Meta{
		PerPage: opts.Pagination.PerPage,
		Page:    opts.Pagination.Page,
		Rows:    count,
	})
}

func (s *Server) GetRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	recurrent := getRecurrentExpenseContext(r)

	s.respond(w, http.StatusOK, recurrent)
}

func (s *Server) PostRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	var params logic.RecurrentExpenseParams

	if err := decodeJSONBody(r, &params); err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	user := getUserContext(r)

	recurrent, err := s.store.CreateRecurrentExpense(r.Context(), user.ID, params)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respond(w, http.StatusCreated, recurrent)
}

func (s *Server) PutRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	var params recurrentExpenseUpdatePayload

	if err := decodeJSONBody(r, &params); err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	recurrent := getRecurrentExpenseContext(r)

	updated, err := s.store.UpdateRecurrentExpense(r.Context(), repo.UpdateRecurrentExpenseParams{
		ID:                recurrent.ID,
		UserID:            recurrent.UserID,
		Description:       params.Description,
		Amount:            params.Amount,
		Period:            params.Period,
		LastCopyCreatedAt: repo.NullInt64FromPtr(recurrent.LastCopyCreatedAt),
	})
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respond(w, http.StatusOK, updated)
}

func (s *Server) DeleteRecurrentExpense(w http.ResponseWriter, r *http.Request) {
	recurrent := getRecurrentExpenseContext(r)

	id, err := s.store.DeleteRecurrentExpense(r.Context(), recurrent.ID, recurrent.UserID)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respond(w, http.StatusOK, map[string]int{"id": id})
}

func getRecurrentExpenseContext(r *http.Request) *repo.RecurrentExpense {
	recurrent, ok := r.Context().Value(prog.KeyRecurrentExpense).(*repo.RecurrentExpense)

	if !ok {
		err := fmt.Errorf("failed to extract recurrent expense context, %w, %v", ErrMissingContext, recurrent)
		panic(err)
	}

	return recurrent
}

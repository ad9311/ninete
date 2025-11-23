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

func (s *Server) ContextExpense(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserContext(r)

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			s.respondError(w, http.StatusNotFound, ErrInvalidId)

			return
		}

		expense, err := s.store.FindExpense(r.Context(), id, user.ID)
		if err != nil {
			if errors.Is(err, logic.ErrNotFound) {
				s.respondError(w, http.StatusNotFound, err)

				return
			}
			s.respondError(w, http.StatusBadRequest, err)

			return
		}

		ctx := context.WithValue(r.Context(), prog.KeyExpense, &expense)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) GetExpenses(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query().Get("options")
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
	if opts.Filters.Connector == "" {
		opts.Filters.Connector = "AND"
	}

	expenses, err := s.store.FindExpenses(r.Context(), opts)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	count, err := s.store.CountExpenses(r.Context(), opts.Filters)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respondMeta(w, http.StatusOK, expenses, Meta{
		PerPage: opts.Pagination.PerPage,
		Page:    opts.Pagination.Page,
		Rows:    count,
	})
}

func (s *Server) GetExpense(w http.ResponseWriter, r *http.Request) {
	expense := getExpenseContext(r)

	s.respond(w, http.StatusOK, expense)
}

func (s *Server) PostExpense(w http.ResponseWriter, r *http.Request) {
	var params logic.ExpenseParams

	if err := decodeJSONBody(r, &params); err != nil {
		s.respondError(w, http.StatusBadRequest, ErrFormParsing)

		return
	}

	user := getUserContext(r)

	expense, err := s.store.CreateExpense(r.Context(), user.ID, params)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respond(w, http.StatusCreated, expense)
}

func (s *Server) PutExpense(w http.ResponseWriter, r *http.Request) {
	var params logic.ExpenseParams

	if err := decodeJSONBody(r, &params); err != nil {
		s.respondError(w, http.StatusBadRequest, ErrFormParsing)

		return
	}

	e := getExpenseContext(r)

	expense, err := s.store.UpdateExpense(r.Context(), e.ID, params)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respond(w, http.StatusOK, expense)
}

func (s *Server) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	e := getExpenseContext(r)

	id, err := s.store.DeleteExpense(r.Context(), e.ID)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respond(w, http.StatusOK, map[string]int{"id": id})
}

func getExpenseContext(r *http.Request) *repo.Expense {
	expense, ok := r.Context().Value(prog.KeyExpense).(*repo.Expense)

	if !ok {
		err := fmt.Errorf("failed to extract expense context, %w, %v", ErrMissingContext, expense)
		panic(err)
	}

	return expense
}

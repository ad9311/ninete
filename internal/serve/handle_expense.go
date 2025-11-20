package serve

import (
	"net/http"

	"github.com/ad9311/ninete/internal/repo"
)

func (s *Server) PostExpense(w http.ResponseWriter, r *http.Request) {
	var params repo.ExpenseParams

	if err := decodeJSONBody(r, &params); err != nil {
		s.respondError(w, http.StatusBadRequest, ErrFormParsing)

		return
	}

	user, ok := getUserContext(r)
	if !ok {
		s.missingUserContext(w)

		return
	}

	expense, err := s.store.CreateExpense(r.Context(), user.ID, params)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respond(w, http.StatusCreated, expense)
}

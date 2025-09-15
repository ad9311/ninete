package serve

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

func (s *Server) postSignUp(w http.ResponseWriter, r *http.Request) {
	var params logic.SignUpParams
	if err := decodeJSONBody(r, &params); err != nil {
		s.respondError(w, http.StatusBadRequest, ErrFormParsing)

		return
	}

	user, err := s.store.SignUpUser(r.Context(), params)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.respond(w, http.StatusCreated, user)
}

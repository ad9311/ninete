package serve

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserContext(r)
	if !ok {
		s.respondError(w, http.StatusInternalServerError, CodeGeneric, logic.ErrNotFound)

		return
	}

	s.respond(w, http.StatusOK, user)
}

func getUserContext(r *http.Request) (*repo.SafeUser, bool) {
	user, ok := r.Context().Value(prog.KeyCurrentUser).(*repo.SafeUser)

	return user, ok && user != nil
}

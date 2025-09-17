package serve

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

func (s *Server) deleteSignOut(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserContext(r)
	if !ok {
		s.respondError(w, http.StatusInternalServerError, logic.ErrNotFound)

		return
	}

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		s.respondError(w, http.StatusUnauthorized, ErrInvalidAuthCreds)

		return
	}

	err = s.store.SignOutUser(r.Context(), user.ID, cookie.Value)
	if err != nil {
		s.respondError(w, http.StatusUnauthorized, ErrInvalidAuthCreds)

		return
	}

	s.respondNoContent(w)
}

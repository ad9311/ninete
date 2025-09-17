package serve

import (
	"fmt"
	"net/http"
)

func (s *Server) deleteSignOut(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserContext(r)
	if !ok {
		s.respondError(
			w,
			http.StatusInternalServerError,
			fmt.Errorf("%w: current user", ErrMissingContext),
		)

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

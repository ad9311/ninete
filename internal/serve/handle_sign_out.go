package serve

import (
	"net/http"
)

func (s *Server) deleteSignOut(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		s.respondNoContent(w)

		return
	}

	_ = s.store.SignOutUser(r.Context(), cookie.Value)

	s.respondNoContent(w)
}

package serve

import (
	"net/http"
)

func (s *Server) deleteSignOut(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie(cookieName)
	_ = s.store.SignOutUser(r.Context(), cookie.Value)

	s.respondNoContent(w)
}

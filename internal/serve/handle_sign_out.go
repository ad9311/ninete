package serve

import (
	"errors"
	"net/http"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
)

func (s *Server) deleteSignOut(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		s.deleteRefreshCookie(w)
		s.respondNoContent(w)

		return
	}

	err = s.store.SignOutUser(r.Context(), cookie.Value)
	if err != nil {
		if !errors.Is(err, logic.ErrNotFound) {
			s.app.Logger.Error("failed to delete refresh token, %v", err)
		}
		s.deleteRefreshCookie(w)
		s.respondNoContent(w)

		return
	}

	s.deleteRefreshCookie(w)
	s.respondNoContent(w)
}

func (s *Server) deleteRefreshCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   s.app.ENV == prog.ENVProduction,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	}

	http.SetCookie(w, cookie)
}

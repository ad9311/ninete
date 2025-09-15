package serve

import (
	"net/http"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

const (
	cookieName = "refresh_token"
	cookiePath = "/auth"
)

type SessionResponse struct {
	User        repo.SafeUser `json:"user"`
	AccessToken logic.Token   `json:"accessToken"`
}

func (s *Server) postSignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var params logic.SessionParams
	if err := decodeJSONBody(r, &params); err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	session, err := s.store.SignInUser(ctx, params)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.setRefreshTokenCookie(w, session)

	res := SessionResponse{
		User:        session.User,
		AccessToken: session.AccessToken,
	}

	s.respond(w, http.StatusCreated, res)
}

func (s *Server) setRefreshTokenCookie(w http.ResponseWriter, session logic.NewSession) {
	secure := s.app.ENV == prog.ENVProduction

	http.SetCookie(w, newRefreshCookie(session.RefreshToken, secure))
}

func newRefreshCookie(token logic.Token, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    token.Value,
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(token.ExpiresAt, 0),
		MaxAge:   int(time.Until(time.Unix(token.ExpiresAt, 0)).Seconds()),
	}
}

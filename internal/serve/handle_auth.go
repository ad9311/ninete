package serve

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ad9311/ninete/internal/logic"
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

func (s *Server) PostSignUp(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) PostSignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var params logic.SessionParams
	if err := decodeJSONBody(r, &params); err != nil {
		s.respondError(w, http.StatusBadRequest, ErrFormParsing)

		return
	}

	session, err := s.store.SignInUser(ctx, params)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)

		return
	}

	s.setRefreshTokenCookie(w, session.RefreshToken)

	res := SessionResponse{
		User:        session.User,
		AccessToken: session.AccessToken,
	}

	s.respond(w, http.StatusCreated, res)
}

func (s *Server) DeleteSignOut(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		s.deleteRefreshCookie(w)
		s.respondNoContent(w)

		return
	}

	err = s.store.SignOutUser(r.Context(), cookie.Value)
	if err != nil {
		if !errors.Is(err, logic.ErrNotFound) {
			s.app.Logger.Errorf("failed to delete refresh token, %v", err)
		}
		s.deleteRefreshCookie(w)
		s.respondNoContent(w)

		return
	}

	s.deleteRefreshCookie(w)
	s.respondNoContent(w)
}

func (s *Server) PostRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		s.respondError(
			w,
			http.StatusUnauthorized,
			fmt.Errorf("%w, refresh cookie not found", ErrInvalidAuthCreds),
		)

		return
	}

	refreshToken, err := s.store.FindRefreshToken(ctx, cookie.Value)
	if err != nil {
		s.respondError(
			w,
			http.StatusUnauthorized,
			fmt.Errorf("%w with refresh token, %w", ErrInvalidAuthCreds, err),
		)

		return
	}

	exp := time.Unix(refreshToken.ExpiresAt, 0)
	if time.Now().UTC().After(exp) {
		s.respondError(
			w,
			http.StatusUnauthorized,
			fmt.Errorf("%w, token has expired", ErrInvalidAuthCreds),
		)

		return
	}

	accessToken, err := s.store.NewAccessToken(refreshToken.UserID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err)

		return
	}

	token := logic.Token{
		Value:     accessToken.Value,
		IssuedAt:  accessToken.IssuedAt,
		ExpiresAt: accessToken.ExpiresAt,
	}

	s.respond(w, http.StatusOK, token)
}

func (s *Server) setRefreshTokenCookie(w http.ResponseWriter, token logic.Token) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token.Value,
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   s.app.IsProduction(),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(token.ExpiresAt, 0),
		MaxAge:   int(time.Until(time.Unix(token.ExpiresAt, 0)).Seconds()),
	})
}

func (s *Server) deleteRefreshCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   s.app.IsProduction(),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	}

	http.SetCookie(w, cookie)
}

package serve

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ad9311/ninete/internal/logic"
)

func (s *Server) postRefresh(w http.ResponseWriter, r *http.Request) {
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

package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
)

// SessionResponse contains the user and access token returned after signing in or refreshing tokens.
type SessionResponse struct {
	User        service.SafeUser `json:"user"`        // Authenticated user information
	AccessToken service.Token    `json:"accessToken"` // JWT access token
}

const (
	cookieName = "refresh_token"
	cookiePath = "/auth"
)

// PostSignIn handles user sign-in, validates credentials, and returns a session response with tokens.
func (s *Server) PostSignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var params service.SessionParams
	if err := decodeJSONBody(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, invalidFormFormatErrorCode, err)

		return
	}

	object, err := s.serviceStore.SignInUser(ctx, params)
	if err != nil {
		if errors.Is(err, errs.ErrValidationFailed) {
			writeError(w, http.StatusBadRequest, invalidFormErrorCode, err)

			return
		}

		writeError(w, http.StatusUnauthorized, standardErrorCode, err)

		return
	}

	s.setRefreshTokenCookie(w, object)

	write(w, http.StatusCreated, SessionResponse{
		User:        object.User,
		AccessToken: object.AccessToken,
	})
}

// PostRefresh refreshes the access token using the refresh token cookie and returns a new session response.
func (s *Server) PostRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, invalidAuthCredsErrorCode, errs.ErrRefreshTokenNotFound)

		return
	}

	refreshToken, err := s.serviceStore.FindRefreshTokenByUUID(ctx, cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, standardErrorCode, err)

		return
	}

	if time.Now().UTC().After(refreshToken.ExpiresAt.Time) {
		writeError(w, http.StatusUnauthorized, standardErrorCode, errs.ErrExpiredRefreshToken)

		return
	}

	if refreshToken.Revoked {
		writeError(w, http.StatusUnauthorized, standardErrorCode, errs.ErrRevokedRefreshToken)

		return
	}

	user, err := s.serviceStore.FindUserByID(ctx, refreshToken.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, standardErrorCode, err)

		return
	}

	accessToken, err := s.serviceStore.GenerateAccessToken(refreshToken.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, standardErrorCode, err)

		return
	}

	write(w, http.StatusCreated, SessionResponse{
		User:        user,
		AccessToken: accessToken,
	})
}

// DeleteSignOut deletes the refresh token, effectively signing out the user.
func (s *Server) DeleteSignOut(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, invalidAuthCredsErrorCode, errs.ErrRefreshTokenNotFound)

		return
	}

	if err := s.serviceStore.SignOutUser(r.Context(), cookie.Value); err != nil {
		writeError(w, http.StatusUnauthorized, standardErrorCode, err)

		return
	}

	writeNoContent(w)
}

// setRefreshTokenCookie sets the refresh token cookie in the response for the signed-in user.
func (s *Server) setRefreshTokenCookie(w http.ResponseWriter, object service.SessionObject) {
	secure := s.config.Env == app.EnvProduction

	http.SetCookie(w, newRefreshCookie(object.RefreshToken, secure))
}

// newRefreshCookie creates a new HTTP cookie for the refresh token with appropriate security settings.
func newRefreshCookie(token service.Token, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    token.Value,
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  token.ExpiresAt,
		MaxAge:   int(time.Until(token.ExpiresAt).Seconds()),
	}
}

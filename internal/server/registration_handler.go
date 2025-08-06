package server

import (
	"errors"
	"net/http"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
)

// PostSignUp handles user registration, validates input, and returns the created user as JSON.
func (s *Server) PostSignUp(w http.ResponseWriter, r *http.Request) {
	var params service.RegistrationParams
	if err := decodeJSONBody(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, invalidFormFormatErrorCode, err)

		return
	}

	user, err := s.serviceStore.RegisterUser(r.Context(), params)
	if err != nil {
		if errors.Is(err, errs.ErrValidationFailed) || errors.Is(err, errs.ErrUnmatchedPasswords) {
			writeError(w, http.StatusBadRequest, invalidFormErrorCode, err)

			return
		}
		writeError(w, http.StatusBadRequest, standardErrorCode, err)

		return
	}

	write(w, http.StatusCreated, service.SafeUser{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

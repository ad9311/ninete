package server

import (
	"net/http"

	"github.com/ad9311/go-api-base/internal/errs"
)

// GetMe handles the /users/me route and returns the current authenticated user as JSON.
func (s *Server) GetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserContext(r)
	if !ok {
		writeError(w, http.StatusInternalServerError, internalErrorCode, errs.ErrNotFound)

		return
	}

	write(w, http.StatusOK, Data{"user": user})
}

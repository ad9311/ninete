package handlers

import "net/http"

// apiSession is the signed-in user as the client sees it. It is a separate type
// from logic.User so a field added there is not published by accident.
type apiSession struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetAPISession answers with the current user. The SPA shell needs it to render
// chrome without a template, and it is the route the API middleware chain is
// verified through — a chi sub-router with no routes at all never builds its
// middleware chain, so the group needs at least one.
func (h *Handler) GetAPISession(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user == nil {
		h.APIUnauthorized(w, r)

		return
	}

	h.WriteJSON(w, http.StatusOK, apiSession{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

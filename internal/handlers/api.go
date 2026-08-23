package handlers

import (
	"encoding/json"
	"net/http"
)

// APIError is the envelope every /api/* failure uses. Fields carries per-field
// validation messages and is absent on everything else — see §3.5 of
// docs/spa-migration.md.
type APIError struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// WriteJSON is the single place an API response body is written. Status goes
// out before the body, so a failure mid-encode can only truncate the payload,
// never change the status the client already read.
func (h *Handler) WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		h.app.Logger.Errorf("failed to write JSON response: %v", err)
	}
}

// WriteJSONError writes err's message as the envelope above. Never pass an
// unexpected internal error here: the message reaches the browser.
func (h *Handler) WriteJSONError(w http.ResponseWriter, status int, err error) {
	h.WriteJSON(w, status, APIError{Error: err.Error()})
}

// APIUnauthorized answers a request that is not signed in. It must never send a
// Location header: fetch follows redirects silently, so the HTML chain's
// 303 → /login would reach the client as the login page's markup under status
// 200, and the JSON parse would fail instead of the auth check.
func (h *Handler) APIUnauthorized(w http.ResponseWriter, _ *http.Request) {
	h.WriteJSONError(w, http.StatusUnauthorized, ErrUnauthorized)
}

// APIForbidden answers a request rejected by CSRF, replacing nosurf's plain
// text default so the client parses a body of the same shape as every other
// failure.
func (h *Handler) APIForbidden(w http.ResponseWriter, _ *http.Request) {
	h.WriteJSONError(w, http.StatusForbidden, ErrInvalidCSRFToken)
}

func (h *Handler) APINotFound(w http.ResponseWriter, _ *http.Request) {
	h.WriteJSONError(w, http.StatusNotFound, ErrAPIRouteNotFound)
}

func (h *Handler) APIMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	h.WriteJSONError(w, http.StatusMethodNotAllowed, ErrNotAllowed)
}

// APIInternalError is the generic fallback. The real error belongs in the log,
// not in the response.
func (h *Handler) APIInternalError(w http.ResponseWriter, _ *http.Request) {
	h.WriteJSONError(w, http.StatusInternalServerError, ErrAPIUnavailable)
}

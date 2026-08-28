package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

// APIError is the envelope every /api/* failure uses. Fields carries per-field
// validation messages and is absent on everything else — see §3.5 of
// docs/spa-migration.md.
type APIError struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// DecodeJSON reads and decodes a JSON request body into dst, answering 400
// with the same envelope every other failure uses when it cannot. The body is
// always closed, matching what net/http does for every other request path.
func (h *Handler) DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer func() {
		_ = r.Body.Close()
	}()

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		h.WriteJSONError(w, http.StatusBadRequest, ErrAPIInvalidJSON)

		return false
	}

	return true
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

// WriteAPIError is the single mapping from a store error to a response. A
// validation failure becomes 422 carrying the per-field rules, a missing row
// becomes 404, and anything not named is treated as a server fault: logged in
// full, answered with a generic message.
//
// userErrors names the errors this endpoint knows are the caller's fault, which
// answer 422 with their own text — the same messages the pages render today.
// Nothing is user-facing by default, so an unexpected driver error cannot
// describe itself to the browser the way a re-rendered form does.
func (h *Handler) WriteAPIError(w http.ResponseWriter, err error, userErrors ...error) {
	var validationErr *logic.ValidationError
	if errors.As(err, &validationErr) {
		h.WriteJSON(w, http.StatusUnprocessableEntity, APIError{
			Error:  validationErr.Error(),
			Fields: validationErr.Fields,
		})

		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		h.WriteJSONError(w, http.StatusNotFound, ErrAPIRouteNotFound)

		return
	}

	for _, userErr := range userErrors {
		if errors.Is(err, userErr) {
			h.WriteJSONError(w, http.StatusUnprocessableEntity, userErr)

			return
		}
	}

	h.app.Logger.Errorf("unhandled API error: %v", err)
	h.WriteJSONError(w, http.StatusInternalServerError, ErrAPIUnavailable)
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

// APITooManyRequests is TooManyRequests' JSON twin, for the credential routes
// on the API chain. It must not reach a render helper: the API chain drops
// setTmplData, and tmplData panics when the template map is absent — so the
// HTML handler answers a throttled /api/login with a recovered panic and an
// empty 500 instead of a 429.
func (h *Handler) APITooManyRequests(w http.ResponseWriter, _ *http.Request) {
	h.WriteJSONError(w, http.StatusTooManyRequests, ErrTooManyAttempts)
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

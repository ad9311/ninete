package handlers

import (
	"io"
	"net/http"
)

// PostCSPReport logs browser-sent CSP violation reports. It gives a
// server-side signal when the nonce-only policy blocks an inline script or
// style, so template regressions surface in logs instead of failing silently
// in the browser. The endpoint is unauthenticated and CSRF-exempt because
// browsers post reports automatically without cookies or tokens.
func (h *Handler) PostCSPReport(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err == nil && len(body) > 0 {
		h.app.Logger.Errorf("csp violation report: %s", body)
	}

	w.WriteHeader(http.StatusNoContent)
}

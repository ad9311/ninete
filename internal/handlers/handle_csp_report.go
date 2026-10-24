package handlers

import (
	"io"
	"net/http"
)

// cspReportMaxBytes is as much of a report as gets logged. A real violation
// report is well under 1 KB; the rest of the 1 MB body cap would only let an
// anonymous caller write a megabyte into the log per request.
const cspReportMaxBytes = 8 << 10

// PostCSPReport logs browser-sent CSP violation reports. It gives a
// server-side signal when the nonce-only policy blocks an inline script or
// style, so a shell regression surfaces in logs instead of failing silently
// in the browser. The endpoint is unauthenticated and CSRF-exempt because
// browsers post reports automatically without cookies or tokens.
//
// Anyone can therefore call it, so the body is cut to cspReportMaxBytes and
// logged quoted: %q escapes newlines and control characters, and a raw body
// containing them could otherwise forge whole log lines. The route also
// carries its own rate limit (cspReportRateLimit in internal/serve).
func (h *Handler) PostCSPReport(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, cspReportMaxBytes))
	if err == nil && len(body) > 0 {
		h.app.Logger.Errorf("csp violation report: %q", body)
	}

	w.WriteHeader(http.StatusNoContent)
}

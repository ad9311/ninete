package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/stretchr/testify/require"
)

// Internal because the log is the behavior under test, and the handler's
// logger is reachable only through Deps.
func TestPostCSPReport(t *testing.T) {
	newHandler := func(out *bytes.Buffer) *Handler {
		logger := prog.NewLogger(prog.LogOptions{Out: out, OutErr: out})

		return New(Deps{App: &prog.App{ENV: prog.ENVTest, Logger: logger}})
	}

	post := func(t *testing.T, h *Handler, body string) int {
		t.Helper()

		req := httptest.NewRequest(http.MethodPost, "/csp-report", strings.NewReader(body))
		rec := httptest.NewRecorder()
		h.PostCSPReport(rec, req)

		return rec.Code
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_log_a_report",
			fn: func(t *testing.T) {
				var out bytes.Buffer

				require.Equal(t, http.StatusNoContent, post(t, newHandler(&out), `{"csp-report":{}}`))
				require.Contains(t, out.String(), `csp-report`)
			},
		},
		{
			// Genuine reproduction: the body used to be logged with %s, so a
			// newline in it started a log line of the caller's choosing.
			name: "should_escape_newlines_so_a_report_cannot_forge_log_lines",
			fn: func(t *testing.T) {
				var out bytes.Buffer

				post(t, newHandler(&out), "x\n[ERROR] forged line")
				require.NotContains(t, out.String(), "\n[ERROR] forged line")
				require.Contains(t, out.String(), `x\n[ERROR] forged line`)
			},
		},
		{
			// Genuine reproduction: the whole body, up to the 1 MB request
			// cap, used to reach the log.
			name: "should_log_at_most_the_report_size_cap",
			fn: func(t *testing.T) {
				var out bytes.Buffer

				post(t, newHandler(&out), strings.Repeat("#", cspReportMaxBytes*4))
				require.Equal(t, cspReportMaxBytes, strings.Count(out.String(), "#"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

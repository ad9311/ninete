package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/report"
)

// reportMonthLayout is the month parameter's format. The report period is one
// calendar month of the billed date, which is stored at month precision and
// carries no zone, so unlike the /api/expenses* ranges (§3.6 of
// docs/spa-migration.md) there is nothing for a client zone to resolve — the
// month is the whole of the input.
const reportMonthLayout = "2006-01"

// GetReportsExpenses answers the monthly expense report as a PDF.
//
// It sits on the *page* chain, not under /api, for the reason
// GetExportsExpenses documents: it is reached by a plain anchor, which is a
// browser navigation, so an expired session has to produce a redirect the
// browser can follow rather than a 401 with no Location.
func (h *Handler) GetReportsExpenses(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)

	month, err := parseReportMonth(r.URL.Query().Get("month"))
	if err != nil {
		h.WriteAPIError(w, err, ErrReportMonth)

		return
	}

	data, err := h.store.BuildMonthlyReport(r.Context(), user.ID, month)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	pdf, err := report.Render(data)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	// Buffered rather than streamed, unlike the JSON export: fpdf builds the
	// whole document before it can write a byte, so there is nothing to gain
	// by withholding Content-Length.
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition",
		`attachment; filename="expense-report-`+month.Format(reportMonthLayout)+`.pdf"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(pdf)))
	w.WriteHeader(http.StatusOK)

	// The bytes are a PDF the renderer produced, served as application/pdf and
	// as an attachment — there is no document for a script to run in.
	if _, err := w.Write(pdf); err != nil { //nolint:gosec // G705: not an HTML response
		h.app.Logger.Errorf("failed to write expense report: %v", err)
	}
}

// parseReportMonth reads the month parameter, defaulting to last month — the
// period the scheduled report will send, and the one a person asking for "the
// report" on the 1st means.
func parseReportMonth(raw string) (time.Time, error) {
	if raw == "" {
		now := time.Now().UTC()

		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0), nil
	}

	month, err := time.Parse(reportMonthLayout, raw)
	if err != nil {
		return time.Time{}, ErrReportMonth
	}

	return month, nil
}

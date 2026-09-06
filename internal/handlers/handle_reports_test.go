package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetReportsExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "pdf_report_user", "pdf_report_user@example.com", "pdf_report_pass")
	category := s.CreateCategory(t, "pdf report category")
	cookies := s.AuthCookies(t, "pdf_report_user@example.com", "pdf_report_pass")

	month := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	s.CreateExpense(t, user.ID, logic.ExpenseParams{
		Description: "pdf report expense",
		Amount:      4200,
		Date:        month.Unix(),
		CategoryID:  category.ID,
	})

	getReport := func(t *testing.T, query string, cookies []*http.Cookie) *httptest.ResponseRecorder {
		t.Helper()

		req := spec.NewGetRequest("/reports/expenses.pdf"+query, cookies)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_a_signed_out_visitor_rather_than_answering_401",
			fn: func(t *testing.T) {
				// The link is a plain anchor, so a navigation is the only thing
				// the browser can act on — a 401 would be saved as a file.
				rec := getReport(t, "", nil)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_answer_a_pdf_attachment",
			fn: func(t *testing.T) {
				rec := getReport(t, "?month=2026-04", cookies)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Equal(t, "application/pdf", rec.Header().Get("Content-Type"))
				require.Equal(t,
					`attachment; filename="expense-report-2026-04.pdf"`,
					rec.Header().Get("Content-Disposition"))
				require.Equal(t, "%PDF", rec.Body.String()[:4])
			},
		},
		{
			name: "should_default_to_last_month",
			fn: func(t *testing.T) {
				now := time.Now().UTC()
				lastMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)

				rec := getReport(t, "", cookies)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Equal(t,
					`attachment; filename="expense-report-`+lastMonth.Format("2006-01")+`.pdf"`,
					rec.Header().Get("Content-Disposition"))
			},
		},
		{
			name: "should_reject_a_malformed_month",
			fn: func(t *testing.T) {
				rec := getReport(t, "?month=april", cookies)

				require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
				require.Contains(t, rec.Body.String(), "YYYY-MM")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

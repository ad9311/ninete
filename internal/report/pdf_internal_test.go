package report

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-pdf/fpdf"
	"github.com/stretchr/testify/require"
)

// newMeasuringPDF builds the minimum fpdf needs to answer GetStringWidth: a
// document with a font selected. truncate measures against the *current* font,
// so the test has to pick the same one the expense rows use.
func newMeasuringPDF(t *testing.T) *fpdf.Fpdf {
	t.Helper()

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 9)

	require.NoError(t, pdf.Error())

	return pdf
}

func TestTruncate(t *testing.T) {
	pdf := newMeasuringPDF(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_leave_a_string_that_already_fits",
			fn: func(t *testing.T) {
				require.Equal(t, "Groceries", truncate(pdf, "Groceries", 60))
			},
		},
		{
			name: "should_leave_an_empty_string",
			fn: func(t *testing.T) {
				require.Empty(t, truncate(pdf, "", 60))
			},
		},
		{
			name: "should_cut_a_long_string_to_a_prefix_plus_an_ellipsis",
			fn: func(t *testing.T) {
				const long = "Annual software licence renewal for the whole toolchain"

				out := truncate(pdf, long, 40)

				require.NotEqual(t, long, out)
				require.True(t, strings.HasSuffix(out, "..."))
				require.True(t, strings.HasPrefix(long, strings.TrimSuffix(out, "...")),
					"the kept part is not a prefix of the input")
			},
		},
		{
			name: "should_produce_something_that_actually_fits",
			fn: func(t *testing.T) {
				// The point of the whole function: fpdf clips silently, so a
				// result still wider than its column would read as a missing
				// word rather than as a cut.
				for _, width := range []float64{10, 20, 40, 80} {
					out := truncate(pdf, "Annual software licence renewal for the whole toolchain", width)
					require.LessOrEqual(t, pdf.GetStringWidth(out), width,
						"result overflows a %.0fmm column", width)
				}
			},
		},
		{
			name: "should_not_cut_a_multibyte_rune_in_half",
			fn: func(t *testing.T) {
				// The loop slices runes, not bytes, and this is the assertion
				// that holds it to that: every character here is two bytes, so
				// a byte-wise shrink lands mid-rune on half its steps. One
				// fixed width would only catch it by luck, so the widths are
				// swept in small increments across the whole string.
				const accented = "Cafetería á é í ó ú ñ üé àè ìò ùç âê îô ûã õ"

				for width := 1.0; width < pdf.GetStringWidth(accented); width += 0.5 {
					out := truncate(pdf, accented, width)

					require.True(t, utf8.ValidString(out),
						"invalid UTF-8 at width %.1f: %q", width, out)
					require.NotContains(t, out, "\ufffd",
						"replacement character at width %.1f", width)
				}
			},
		},
		{
			name: "should_drop_the_ellipsis_when_not_even_one_character_fits_with_it",
			fn: func(t *testing.T) {
				// A width this small cannot hold a character plus "...", so the
				// loop runs out and returns a bare leading rune. One character
				// is a better answer than an ellipsis wider than the column.
				out := truncate(pdf, "Groceries", 0.5)

				require.Equal(t, "G", out)
			},
		},
		{
			name: "should_return_a_single_rune_unchanged_when_it_cannot_fit",
			fn: func(t *testing.T) {
				// Nothing to shrink: the loop never runs.
				require.Equal(t, "W", truncate(pdf, "W", 0.1))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

package report

import (
	"strconv"
	"strings"
)

// FormatCents renders an unsigned cent amount the way the SPA's
// lib/currency.ts does — "$1,234.56". The browser gets this from Intl; Go has
// no thousands separator in the standard library, so the grouping is done by
// hand here rather than pulling in golang.org/x/text for one column of a PDF.
func FormatCents(cents uint64) string {
	return "$" + groupThousands(cents/centsPerUnit) + "." + pad2(cents%centsPerUnit)
}

// FormatSignedCents is FormatCents for a value that may be negative — a
// budget's remaining amount once it has been exceeded. The sign leads the
// currency symbol, matching how Intl prints "-$12.34".
func FormatSignedCents(cents int64) string {
	if cents < 0 {
		return "-" + FormatCents(uint64(-cents)) //nolint:gosec // a cent amount, nowhere near the int64 bound
	}

	return FormatCents(uint64(cents)) //nolint:gosec // guarded non-negative above
}

const centsPerUnit = 100

func pad2(n uint64) string {
	if n < 10 {
		return "0" + strconv.FormatUint(n, 10)
	}

	return strconv.FormatUint(n, 10)
}

func groupThousands(n uint64) string {
	digits := strconv.FormatUint(n, 10)
	if len(digits) <= 3 {
		return digits
	}

	var b strings.Builder

	lead := len(digits) % 3
	if lead > 0 {
		b.WriteString(digits[:lead])
	}

	for i := lead; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}

		b.WriteString(digits[i : i+3])
	}

	return b.String()
}

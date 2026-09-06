package report_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/report"
	"github.com/stretchr/testify/require"
)

func TestFormatCents(t *testing.T) {
	cases := []struct {
		name  string
		cents uint64
		want  string
	}{
		{name: "zero", cents: 0, want: "$0.00"},
		{name: "cents_only", cents: 7, want: "$0.07"},
		{name: "under_a_thousand", cents: 99999, want: "$999.99"},
		{name: "one_separator", cents: 123456, want: "$1,234.56"},
		{name: "exactly_a_thousand", cents: 100000, want: "$1,000.00"},
		{name: "two_separators", cents: 123456789, want: "$1,234,567.89"},
		{name: "leading_group_of_one", cents: 100000000, want: "$1,000,000.00"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, report.FormatCents(c.cents))
		})
	}
}

func TestFormatSignedCents(t *testing.T) {
	cases := []struct {
		name  string
		cents int64
		want  string
	}{
		{name: "positive", cents: 123456, want: "$1,234.56"},
		{name: "zero", cents: 0, want: "$0.00"},
		{name: "negative", cents: -123456, want: "-$1,234.56"},
		{name: "negative_cents_only", cents: -5, want: "-$0.05"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, report.FormatSignedCents(c.cents))
		})
	}
}

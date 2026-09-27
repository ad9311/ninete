package units_test

import (
	"testing"
	"time"

	"github.com/ad9311/ninete-mcp/internal/units"
	"github.com/stretchr/testify/require"
)

func TestUnits(t *testing.T) {
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)

	losAngeles, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_parse_amounts_into_cents",
			fn: func(t *testing.T) {
				valid := map[string]uint64{
					"12":      1200,
					"12.5":    1250,
					"12.50":   1250,
					"0.07":    7,
					"0":       0,
					"3.10":    310,
					"1234.56": 123456,
				}

				for raw, want := range valid {
					got, err := units.ParseAmount(raw)
					require.NoError(t, err, raw)
					require.Equal(t, want, got, raw)
				}

				for _, raw := range []string{
					"", "-1", "1.234", "1.", ".5", "1,000", "abc", "1e3", "+2",
					// Past uint64 cents: the multiplication would wrap to a small amount.
					"184467440737095517", "18446744073709551615",
				} {
					_, err := units.ParseAmount(raw)
					require.ErrorIs(t, err, units.ErrAmount, raw)
				}
			},
		},
		{
			name: "should_format_cents_without_floats",
			fn: func(t *testing.T) {
				require.Equal(t, "12.50", units.FormatAmount(1250))
				require.Equal(t, "0.07", units.FormatAmount(7))
				require.Equal(t, "-3.05", units.FormatSignedAmount(-305))
				require.Equal(t, "3.05", units.FormatSignedAmount(305))
			},
		},
		{
			name: "should_round_trip_billed_months_in_utc",
			fn: func(t *testing.T) {
				unix, err := units.ParseMonth("2026-09")
				require.NoError(t, err)
				require.Equal(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Unix(), unix)
				require.Equal(t, "2026-09", units.FormatMonth(unix))

				_, err = units.ParseMonth("2026-13")
				require.ErrorIs(t, err, units.ErrMonth)
			},
		},
		{
			name: "should_resolve_inclusive_month_ranges",
			fn: func(t *testing.T) {
				start, end, err := units.MonthRange("2026-11", "2027-01")
				require.NoError(t, err)
				require.Equal(t, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Unix(), start)
				require.Equal(t, time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC).Unix(), end)

				start, end, err = units.MonthRange("2026-09", "")
				require.NoError(t, err)
				require.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).Unix(), end)
				require.Less(t, start, end)

				_, _, err = units.MonthRange("2026-09", "2026-08")
				require.ErrorIs(t, err, units.ErrRange)
			},
		},
		{
			name: "should_resolve_day_ranges_in_the_given_zone",
			fn: func(t *testing.T) {
				start, end, err := units.DayRange("2026-09-26", "", auckland)
				require.NoError(t, err)
				require.Equal(t, time.Date(2026, 9, 26, 0, 0, 0, 0, auckland).Unix(), start)
				require.Equal(t, time.Date(2026, 9, 27, 0, 0, 0, 0, auckland).Unix(), end)

				// 2026-11-01 is the US fall-back day: 25 hours long.
				start, end, err = units.DayRange("2026-11-01", "2026-11-01", losAngeles)
				require.NoError(t, err)
				require.Equal(t, int64(25*60*60), end-start)

				_, _, err = units.DayRange("26/09/2026", "", auckland)
				require.ErrorIs(t, err, units.ErrDay)
			},
		},
		{
			name: "should_read_the_current_month_in_the_zone",
			fn: func(t *testing.T) {
				// 2026-09-30 20:00 UTC is already October in Auckland.
				now := time.Date(2026, 9, 30, 20, 0, 0, 0, time.UTC)
				require.Equal(t, "2026-10", units.CurrentMonth(now, auckland))
				require.Equal(t, "2026-09", units.CurrentMonth(now, losAngeles))

				prev, err := units.PreviousMonth("2026-01")
				require.NoError(t, err)
				require.Equal(t, "2025-12", prev)
			},
		},
		{
			name: "should_match_javascript_timezone_offset_sign",
			fn: func(t *testing.T) {
				summer := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
				require.Equal(t, 420, units.TZOffsetMinutes(summer, losAngeles))
				require.Equal(t, -720, units.TZOffsetMinutes(summer, auckland))
				require.Equal(t, 0, units.TZOffsetMinutes(summer, time.UTC))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

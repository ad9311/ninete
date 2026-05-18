package handlers

import (
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestComputeMacroProgress(t *testing.T) {
	cases := []struct {
		name   string
		totals repo.MacroDayTotals
		goal   repo.MacroGoal
		want   macroProgressData
	}{
		{
			name:   "should_return_zero_when_all_goals_are_zero",
			totals: repo.MacroDayTotals{Kcal: 100, ProteinG: 10, CarbsG: 20, FatG: 5},
			goal:   repo.MacroGoal{},
			want:   macroProgressData{},
		},
		{
			name:   "should_compute_partial_progress",
			totals: repo.MacroDayTotals{Kcal: 500, ProteinG: 50, CarbsG: 100, FatG: 25},
			goal:   repo.MacroGoal{Kcal: 2000, ProteinG: 100, CarbsG: 200, FatG: 50},
			want: macroProgressData{
				KcalPct: 25, ProteinPct: 50, CarbsPct: 50, FatPct: 50,
			},
		},
		{
			name:   "should_cap_at_100_when_totals_exceed_goal",
			totals: repo.MacroDayTotals{Kcal: 3000, ProteinG: 200, CarbsG: 400, FatG: 80},
			goal:   repo.MacroGoal{Kcal: 2000, ProteinG: 100, CarbsG: 200, FatG: 50},
			want: macroProgressData{
				KcalPct: 100, ProteinPct: 100, CarbsPct: 100, FatPct: 100,
			},
		},
		{
			name:   "should_handle_negative_goal_as_zero_progress",
			totals: repo.MacroDayTotals{Kcal: 100},
			goal:   repo.MacroGoal{Kcal: -50},
			want:   macroProgressData{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeMacroProgress(tc.totals, tc.goal)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestComputeMacroTrendSummary(t *testing.T) {
	cases := []struct {
		name string
		in   []repo.MacroDailyTotal
		want macroTrendSummary
	}{
		{
			name: "should_return_zeros_for_empty_input",
			in:   nil,
			want: macroTrendSummary{},
		},
		{
			name: "should_sum_and_average_across_days",
			in: []repo.MacroDailyTotal{
				{Kcal: 2000, ProteinG: 100, CarbsG: 200, FatG: 50},
				{Kcal: 2400, ProteinG: 120, CarbsG: 240, FatG: 70},
			},
			want: macroTrendSummary{
				ActiveDays:   2,
				TotalKcal:    4400,
				AvgKcal:      2200,
				TotalProtein: 220,
				AvgProtein:   110,
				TotalCarbs:   440,
				AvgCarbs:     220,
				TotalFat:     120,
				AvgFat:       60,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeMacroTrendSummary(tc.in)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestComputeDayWindow(t *testing.T) {
	cases := []struct {
		name             string
		in               string
		wantSelected     string
		wantNextIsPlus1d bool
	}{
		{
			name:             "should_use_today_when_input_is_empty",
			in:               "",
			wantSelected:     "", // computed at runtime
			wantNextIsPlus1d: true,
		},
		{
			name:             "should_use_today_when_input_is_malformed",
			in:               "not-a-date",
			wantSelected:     "",
			wantNextIsPlus1d: true,
		},
		{
			name:             "should_parse_iso_date",
			in:               "2026-03-15",
			wantSelected:     "2026-03-15",
			wantNextIsPlus1d: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dayStart, nextDayStart, selectedDate := computeDayWindow(tc.in)

			require.Equal(t, int64(86400), nextDayStart-dayStart, "next day should be exactly 24h after dayStart")
			if tc.wantSelected != "" {
				require.Equal(t, tc.wantSelected, selectedDate)
			} else {
				require.Len(t, selectedDate, 10, "selectedDate should be YYYY-MM-DD")
			}
		})
	}
}

func TestNextSortOrder(t *testing.T) {
	cases := []struct {
		name                                              string
		currentField, currentOrder, column, columnDefault string
		want                                              string
	}{
		{
			"should_use_column_default_when_not_currently_sorting_by_column",
			"count", "DESC", "mood", "ASC", "ASC",
		},
		{
			"should_flip_asc_to_desc_when_currently_sorting_by_column_asc",
			"mood", "ASC", "mood", "ASC", "DESC",
		},
		{
			"should_flip_desc_to_asc_when_currently_sorting_by_column_desc",
			"mood", "DESC", "mood", "ASC", "ASC",
		},
		{
			"should_use_count_default_when_current_field_is_mood",
			"mood", "DESC", "count", "DESC", "DESC",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := nextSortOrder(tc.currentField, tc.currentOrder, tc.column, tc.columnDefault)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestRoundMacro(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want float64
	}{
		{"should_round_to_two_decimals", 1.23456, 1.23},
		{"should_round_half_up", 1.235, 1.24},
		{"should_pass_through_whole_numbers", 100.0, 100.0},
		{"should_handle_zero", 0, 0},
		{"should_round_negative_values_away_from_zero", -1.235, -1.24},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.InDelta(t, tc.want, roundMacro(tc.in), 0.0001)
		})
	}
}

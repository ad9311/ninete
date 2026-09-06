package logic_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestParseQuickExpense(t *testing.T) {
	// The relative keywords resolve against the client's calendar month; these
	// cases pass tzOffsetMinutes=0, so the expectation is UTC's.
	utcFirstOfMonth := func(offset int) int64 {
		now := time.Now().UTC()

		return time.Date(now.Year(), now.Month()+time.Month(offset), 1, 0, 0, 0, 0, time.UTC).Unix()
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_parse_decimal_amount_and_current_month",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 3344.22, current", 0)
				require.NoError(t, err)
				require.Equal(t, "Uber", parsed.Description)
				require.Equal(t, uint64(334422), parsed.Amount)
				require.Equal(t, utcFirstOfMonth(0), parsed.Date)
			},
		},
		{
			name: "should_parse_integer_amount_as_whole_dollars",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Rent, 23044, current", 0)
				require.NoError(t, err)
				require.Equal(t, uint64(2304400), parsed.Amount)
			},
		},
		{
			name: "should_parse_small_decimal_amount",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Coffee, 33.33, current", 0)
				require.NoError(t, err)
				require.Equal(t, uint64(3333), parsed.Amount)
			},
		},
		{
			name: "should_parse_last",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, last", 0)
				require.NoError(t, err)
				require.Equal(t, utcFirstOfMonth(-1), parsed.Date)
			},
		},
		{
			name: "should_parse_explicit_lowercase_long_month",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, june 2026", 0)
				require.NoError(t, err)
				expected := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC).Unix()
				require.Equal(t, expected, parsed.Date)
			},
		},
		{
			name: "should_parse_explicit_short_month",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, Jun 2026", 0)
				require.NoError(t, err)
				expected := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC).Unix()
				require.Equal(t, expected, parsed.Date)
			},
		},
		{
			name: "should_parse_iso_month",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, 2026-06", 0)
				require.NoError(t, err)
				expected := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC).Unix()
				require.Equal(t, expected, parsed.Date)
			},
		},
		{
			name: "should_resolve_relative_months_against_the_client_zone",
			fn: func(t *testing.T) {
				// Only meaningful around a month turnover, so this asserts the
				// offset is applied at all rather than a specific month: an
				// offset a full day wide moves "current" for most of the month.
				utc, err := logic.ParseQuickExpense("Uber, 10, current", 0)
				require.NoError(t, err)

				now := time.Now().UTC()
				// Shifting the client back by today's day-of-month lands it in
				// the previous month whatever today is: day N minus N days is
				// day 0, which normalizes to the last day of the month before.
				minutes := now.Day() * 24 * 60
				shifted, err := logic.ParseQuickExpense("Uber, 10, current", minutes)
				require.NoError(t, err)

				require.Equal(t, utcFirstOfMonth(0), utc.Date)
				require.Equal(t, utcFirstOfMonth(-1), shifted.Date)
			},
		},
		{
			name: "should_fail_on_wrong_field_count",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, 10", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseFormat)
			},
		},
		{
			name: "should_fail_on_invalid_amount",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, abc, current", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseAmount)
			},
		},
		{
			name: "should_fail_on_invalid_date",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, 10, someday", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseDate)
			},
		},
		{
			name: "should_parse_next",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Rent, 500, next", 0)
				require.NoError(t, err)
				require.Equal(t, utcFirstOfMonth(1), parsed.Date)
			},
		},
		{
			// The day-level grammar quick-add used to accept. These all parsed
			// before the month-only change and must not creep back in:
			// expenses.date is a billed month, so a day the user typed here
			// would be silently discarded by everything downstream.
			name: "should_reject_the_retired_day_level_forms",
			fn: func(t *testing.T) {
				for _, input := range []string{
					"today", "yesterday", "tomorrow", "next month",
					"12 July 2026", "12 Jul 2026", "2026-07-12", "12/07/2026", "12-07-2026",
				} {
					_, err := logic.ParseQuickExpense("Uber, 10, "+input, 0)
					require.ErrorIsf(t, err, logic.ErrQuickExpenseDate, "input %q", input)
				}
			},
		},
		{
			name: "should_fail_on_short_description_before_amount_or_date",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("ab, 10, current", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseDescription)
			},
		},
		{
			name: "should_fail_on_long_description",
			fn: func(t *testing.T) {
				long := strings.Repeat("a", 51)
				_, err := logic.ParseQuickExpense(long+", 10, current", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseDescription)
			},
		},
		{
			name: "should_fail_on_zero_amount",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, 0, current", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseAmount)
			},
		},
		{
			name: "should_fail_on_amount_overflowing_cents",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, 1e18, current", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseAmount)
			},
		},
		{
			name: "should_parse_no_tags_when_field_absent",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, current", 0)
				require.NoError(t, err)
				require.Empty(t, parsed.Tags)
			},
		},
		{
			name: "should_parse_semicolon_separated_tags",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, current, MyTag1; mytag2 ", 0)
				require.NoError(t, err)
				require.Equal(t, []string{"mytag1", "mytag2"}, parsed.Tags)
			},
		},
		{
			name: "should_parse_single_tag",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, current, travel", 0)
				require.NoError(t, err)
				require.Equal(t, []string{"travel"}, parsed.Tags)
			},
		},
		{
			name: "should_drop_empty_and_duplicate_tags",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, current, travel;;TRAVEL; work", 0)
				require.NoError(t, err)
				require.Equal(t, []string{"travel", "work"}, parsed.Tags)
			},
		},
		{
			name: "should_parse_empty_tag_field_as_no_tags",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, current,   ", 0)
				require.NoError(t, err)
				require.Empty(t, parsed.Tags)
			},
		},
		{
			name: "should_fail_on_too_many_fields",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, 10, current, tag, extra", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseFormat)
			},
		},
		{
			name: "should_fail_on_too_many_tags",
			fn: func(t *testing.T) {
				tags := make([]string, 11)
				for i := range tags {
					tags[i] = "tag" + strconv.Itoa(i)
				}
				_, err := logic.ParseQuickExpense("Uber, 10, current, "+strings.Join(tags, ";"), 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseTags)
			},
		},
		{
			name: "should_fail_on_long_tag_name",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, 10, current, "+strings.Repeat("a", 21), 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseTagName)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestQuickExpenseCategoryMapping(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "quick_user_1",
		Email:        "quick_user_1@example.com",
		PasswordHash: []byte("quick_user_hash_1"),
	})
	category := s.CreateCategory(t, "quick category 1")
	other := s.CreateCategory(t, "quick category 2")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_report_no_mapping_before_first_use",
			fn: func(t *testing.T) {
				_, found, err := s.Store.ResolveQuickExpenseCategory(ctx, user.ID, "Netflix")
				require.NoError(t, err)
				require.False(t, found)
			},
		},
		{
			name: "should_create_expense_and_remember_category",
			fn: func(t *testing.T) {
				parsed := logic.QuickExpenseParsed{
					Description: "Netflix",
					Amount:      1599,
					Date:        1735689600,
				}
				expense, err := s.Store.CreateQuickExpense(ctx, user.ID, category.ID, parsed)
				require.NoError(t, err)
				require.Positive(t, expense.ID)
				require.Equal(t, category.ID, expense.CategoryID)

				id, found, err := s.Store.ResolveQuickExpenseCategory(ctx, user.ID, "netflix")
				require.NoError(t, err)
				require.True(t, found)
				require.Equal(t, category.ID, id)
			},
		},
		{
			name: "should_overwrite_remembered_category_on_reuse",
			fn: func(t *testing.T) {
				parsed := logic.QuickExpenseParsed{
					Description: "Spotify",
					Amount:      999,
					Date:        1735689600,
				}
				_, err := s.Store.CreateQuickExpense(ctx, user.ID, category.ID, parsed)
				require.NoError(t, err)

				_, err = s.Store.CreateQuickExpense(ctx, user.ID, other.ID, parsed)
				require.NoError(t, err)

				id, found, err := s.Store.ResolveQuickExpenseCategory(ctx, user.ID, "Spotify")
				require.NoError(t, err)
				require.True(t, found)
				require.Equal(t, other.ID, id)
			},
		},
		{
			name: "should_attach_parsed_tags_to_expense",
			fn: func(t *testing.T) {
				parsed := logic.QuickExpenseParsed{
					Description: "Disney Plus",
					Amount:      1299,
					Date:        1735689600,
					Tags:        []string{"streaming", "monthly"},
				}
				expense, err := s.Store.CreateQuickExpense(ctx, user.ID, category.ID, parsed)
				require.NoError(t, err)

				tags, err := s.Store.FindExpenseTags(ctx, expense.ID, user.ID)
				require.NoError(t, err)
				require.ElementsMatch(t, []string{"streaming", "monthly"}, logic.ExtractTagNames(tags))
			},
		},
		{
			name: "should_fail_validation_for_short_description",
			fn: func(t *testing.T) {
				parsed := logic.QuickExpenseParsed{Description: "no", Amount: 100, Date: 1735689600}
				_, err := s.Store.CreateQuickExpense(ctx, user.ID, category.ID, parsed)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

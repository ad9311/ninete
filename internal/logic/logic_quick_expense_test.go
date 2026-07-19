package logic_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestParseQuickExpense(t *testing.T) {
	utcMidnightToday := func() int64 {
		y, m, d := time.Now().UTC().Date()

		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix()
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_parse_decimal_amount_and_today",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 3344.22, today", 0)
				require.NoError(t, err)
				require.Equal(t, "Uber", parsed.Description)
				require.Equal(t, uint64(334422), parsed.Amount)
				require.Equal(t, utcMidnightToday(), parsed.Date)
			},
		},
		{
			name: "should_parse_integer_amount_as_whole_dollars",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Rent, 23044, today", 0)
				require.NoError(t, err)
				require.Equal(t, uint64(2304400), parsed.Amount)
			},
		},
		{
			name: "should_parse_small_decimal_amount",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Coffee, 33.33, today", 0)
				require.NoError(t, err)
				require.Equal(t, uint64(3333), parsed.Amount)
			},
		},
		{
			name: "should_parse_yesterday",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, yesterday", 0)
				require.NoError(t, err)
				require.Equal(t, utcMidnightToday()-int64((time.Hour*24).Seconds()), parsed.Date)
			},
		},
		{
			name: "should_parse_explicit_lowercase_month_date",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, 12 june 2026", 0)
				require.NoError(t, err)
				expected := time.Date(2026, time.June, 12, 0, 0, 0, 0, time.UTC).Unix()
				require.Equal(t, expected, parsed.Date)
			},
		},
		{
			name: "should_parse_iso_date",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, 2026-06-12", 0)
				require.NoError(t, err)
				expected := time.Date(2026, time.June, 12, 0, 0, 0, 0, time.UTC).Unix()
				require.Equal(t, expected, parsed.Date)
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
				_, err := logic.ParseQuickExpense("Uber, abc, today", 0)
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
			name: "should_parse_tomorrow",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Uber, 10, tomorrow", 0)
				require.NoError(t, err)
				now := time.Now().UTC()
				want := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
				require.Equal(t, want.Unix(), parsed.Date)
			},
		},
		{
			name: "should_parse_next_month_as_first_day",
			fn: func(t *testing.T) {
				parsed, err := logic.ParseQuickExpense("Rent, 500, next month", 0)
				require.NoError(t, err)
				now := time.Now().UTC()
				want := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
				require.Equal(t, want.Unix(), parsed.Date)
			},
		},
		{
			name: "should_fail_on_short_description_before_amount_or_date",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("ab, 10, today", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseDescription)
			},
		},
		{
			name: "should_fail_on_long_description",
			fn: func(t *testing.T) {
				long := strings.Repeat("a", 51)
				_, err := logic.ParseQuickExpense(long+", 10, today", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseDescription)
			},
		},
		{
			name: "should_fail_on_zero_amount",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, 0, today", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseAmount)
			},
		},
		{
			name: "should_fail_on_amount_overflowing_cents",
			fn: func(t *testing.T) {
				_, err := logic.ParseQuickExpense("Uber, 1e18, today", 0)
				require.ErrorIs(t, err, logic.ErrQuickExpenseAmount)
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

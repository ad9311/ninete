package logic

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ad9311/ninete/internal/repo"
)

// Quick-add description bounds mirror ExpenseBaseParams.Description validation
// (min=3,max=50) so the input is rejected before any category prompt.
const (
	quickDescriptionMin = 3
	quickDescriptionMax = 50
	// quickMaxDollars caps the amount so dollars*100 stays within uint64 cents.
	quickMaxDollars = float64(math.MaxInt64) / 100
)

// quickDateLayouts are attempted in order when parsing an explicit date.
var quickDateLayouts = []string{ //nolint:gochecknoglobals // static lookup table
	"2 January 2006",
	"2 Jan 2006",
	"January 2 2006",
	"Jan 2 2006",
	"2006-01-02",
	"02/01/2006",
	"02-01-2006",
}

// QuickExpenseParsed holds the fields extracted from a quick-add input string.
type QuickExpenseParsed struct {
	Description string
	Amount      uint64
	Date        int64
}

// ParseQuickExpense parses a "description, amount, date" input into structured
// fields. tzOffsetMinutes is the client's Date.getTimezoneOffset() value, used
// to resolve relative dates ("today"/"yesterday") to the client's calendar day.
func ParseQuickExpense(raw string, tzOffsetMinutes int) (QuickExpenseParsed, error) {
	var parsed QuickExpenseParsed

	parts := strings.Split(raw, ",")
	if len(parts) != 3 {
		return parsed, ErrQuickExpenseFormat
	}

	description := strings.TrimSpace(parts[0])
	if length := utf8.RuneCountInString(description); length < quickDescriptionMin || length > quickDescriptionMax {
		return parsed, ErrQuickExpenseDescription
	}

	amount, err := parseDollarsToCents(strings.TrimSpace(parts[1]))
	if err != nil {
		return parsed, err
	}

	date, err := parseQuickDate(strings.TrimSpace(parts[2]), tzOffsetMinutes)
	if err != nil {
		return parsed, err
	}

	parsed.Description = description
	parsed.Amount = amount
	parsed.Date = date

	return parsed, nil
}

func parseDollarsToCents(s string) (uint64, error) {
	if s == "" {
		return 0, ErrQuickExpenseAmount
	}

	// Amounts are plain decimals (e.g. "3344.22"); no thousands separators —
	// the comma is the field separator, so a comma here never reaches parsing.
	dollars, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrQuickExpenseAmount, err)
	}
	if dollars <= 0 || math.IsInf(dollars, 0) || math.IsNaN(dollars) || dollars > quickMaxDollars {
		return 0, ErrQuickExpenseAmount
	}

	return uint64(math.Round(dollars * 100)), nil
}

func parseQuickDate(s string, tzOffsetMinutes int) (int64, error) {
	loc := time.FixedZone("client", -tzOffsetMinutes*60)
	now := time.Now().In(loc)
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	switch strings.ToLower(s) {
	case "today":
		return today.Unix(), nil
	case "yesterday":
		return today.AddDate(0, 0, -1).Unix(), nil
	}

	titled := titleCaseWords(s)
	for _, layout := range quickDateLayouts {
		if t, err := time.Parse(layout, titled); err == nil {
			return t.Unix(), nil
		}
	}

	return 0, ErrQuickExpenseDate
}

// titleCaseWords upper-cases the first rune of each whitespace-separated word so
// month names like "june" match Go's "January" reference layout.
func titleCaseWords(s string) string {
	fields := strings.Fields(s)
	for i, field := range fields {
		runes := []rune(strings.ToLower(field))
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		fields[i] = string(runes)
	}

	return strings.Join(fields, " ")
}

// ResolveQuickExpenseCategory looks up the remembered category for a user's
// description. The bool return is false when no mapping exists yet.
func (s *Store) ResolveQuickExpenseCategory(
	ctx context.Context,
	userID int,
	description string,
) (int, bool, error) {
	mapping, found, err := s.queries.SelectExpenseCategoryMapping(ctx, userID, descriptionKey(description))
	if err != nil {
		return 0, false, err
	}
	if !found {
		return 0, false, nil
	}

	return mapping.CategoryID, true, nil
}

// CreateQuickExpense creates an expense from parsed quick-add fields and, in the
// same transaction, remembers the description-to-category mapping for reuse.
func (s *Store) CreateQuickExpense(
	ctx context.Context,
	userID, categoryID int,
	parsed QuickExpenseParsed,
) (repo.Expense, error) {
	var expense repo.Expense

	params := ExpenseParams{
		ExpenseBaseParams: ExpenseBaseParams{
			CategoryID:  categoryID,
			Description: parsed.Description,
			Amount:      parsed.Amount,
		},
		Date: parsed.Date,
	}
	if err := s.ValidateStruct(params); err != nil {
		return expense, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		expense, txErr = tq.InsertExpense(ctx, repo.InsertExpenseParams{
			UserID:      userID,
			CategoryID:  params.CategoryID,
			Description: params.Description,
			Amount:      params.Amount,
			Date:        params.Date,
		})
		if txErr != nil {
			return txErr
		}

		_, txErr = tq.UpsertExpenseCategoryMapping(ctx, repo.UpsertExpenseCategoryMappingParams{
			UserID:         userID,
			CategoryID:     categoryID,
			DescriptionKey: descriptionKey(parsed.Description),
		})

		return txErr
	})
	if err != nil {
		return expense, err
	}

	return expense, nil
}

func descriptionKey(description string) string {
	return strings.ToLower(strings.TrimSpace(description))
}

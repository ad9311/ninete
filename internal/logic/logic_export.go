package logic

import (
	"context"
	"sort"

	"github.com/ad9311/ninete/internal/repo"
)

type ExportCategory struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type ExportExpense struct {
	ID          int             `json:"id"`
	Description string          `json:"description"`
	Amount      uint64          `json:"amount"`
	BilledAt    int64           `json:"billed_at"`
	CreatedAt   int64           `json:"created_at"`
	UpdatedAt   int64           `json:"updated_at"`
	Category    *ExportCategory `json:"category"`
	Tags        []string        `json:"tags"`
}

func (s *Store) ExportExpenses(ctx context.Context, userID int) ([]ExportExpense, error) {
	opts := repo.QueryOptions{
		Sorting: repo.Sorting{Field: "date", Order: "DESC"},
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{Name: "user_id", Value: userID, Operator: "="},
			},
			Connector: "AND",
		},
	}

	expenses, err := s.queries.SelectExpenses(ctx, opts)
	if err != nil {
		return nil, err
	}

	categories, err := s.queries.SelectCategories(ctx)
	if err != nil {
		return nil, err
	}
	categoryByID := make(map[int]repo.Category, len(categories))
	for _, c := range categories {
		categoryByID[c.ID] = c
	}

	expenseIDs := make([]int, 0, len(expenses))
	for _, e := range expenses {
		expenseIDs = append(expenseIDs, e.ID)
	}

	tagRows, err := s.queries.SelectTagRows(ctx, repo.TaggableTypeExpense, "expenses", expenseIDs, userID)
	if err != nil {
		return nil, err
	}

	tagsByExpenseID := map[int][]string{}
	for _, row := range tagRows {
		tagsByExpenseID[row.TargetID] = append(tagsByExpenseID[row.TargetID], row.TagName)
	}
	for id := range tagsByExpenseID {
		sort.Strings(tagsByExpenseID[id])
	}

	out := make([]ExportExpense, 0, len(expenses))
	for _, e := range expenses {
		var cat *ExportCategory
		if c, ok := categoryByID[e.CategoryID]; ok {
			cat = &ExportCategory{Name: c.Name, UID: c.UID}
		}

		tags := tagsByExpenseID[e.ID]
		if tags == nil {
			tags = []string{}
		}

		out = append(out, ExportExpense{
			ID:          e.ID,
			Description: e.Description,
			Amount:      e.Amount,
			BilledAt:    e.Date,
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
			Category:    cat,
			Tags:        tags,
		})
	}

	return out, nil
}

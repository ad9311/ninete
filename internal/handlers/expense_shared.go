package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

type expenseFormBase struct {
	CategoryID  int
	Description string
	Amount      uint64
}

func parseExpenseFormBase(r *http.Request) (expenseFormBase, error) {
	var base expenseFormBase

	if err := r.ParseForm(); err != nil {
		return base, fmt.Errorf("failed to parse form, %w", err)
	}

	categoryID, err := prog.ParseID(r.FormValue("category_id"), "Category ID")
	if err != nil {
		return base, err
	}

	amount, err := prog.ParseAmount(r.FormValue("amount"))
	if err != nil {
		return base, err
	}

	base.CategoryID = categoryID
	base.Description = r.FormValue("description")
	base.Amount = amount

	return base, nil
}

func (h *Handler) findCategories(
	ctx context.Context,
) ([]repo.Category, map[int]string, error) {
	categories, err := h.store.FindCategories(ctx)
	if err != nil {
		return categories, nil, err
	}

	categoryNameByID := make(map[int]string, len(categories))
	for _, category := range categories {
		categoryNameByID[category.ID] = category.Name
	}

	return categories, categoryNameByID, nil
}

func setResourceFormData(
	data map[string]any,
	categories []repo.Category,
	resourceName string,
	resource any,
) {
	data["categories"] = categories
	data[resourceName] = resource
}

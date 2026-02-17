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

func userScopedQueryOpts(r *http.Request, userID int) repo.QueryOptions {
	q := r.URL.Query()

	opts := repo.QueryOptions{
		Sorting: repo.Sorting{
			Order: q.Get("sort_order"),
			Field: q.Get("sort_field"),
		},
	}
	opts.Filters.FilterFields = append(opts.Filters.FilterFields, repo.FilterField{
		Name:     "user_id",
		Value:    userID,
		Operator: "=",
	})
	opts.Filters.Connector = "AND"

	return opts
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

func (h *Handler) findCategoriesOrErr(
	w http.ResponseWriter,
	r *http.Request,
	tmpl TemplateName,
) ([]repo.Category, map[int]string, bool) {
	categories, nameByID, err := h.findCategories(r.Context())
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, tmpl, err)

		return nil, nil, false
	}

	return categories, nameByID, true
}

func categoryNameOrUnknown(nameByID map[int]string, categoryID int) string {
	if name := nameByID[categoryID]; name != "" {
		return name
	}

	return "Unknown"
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

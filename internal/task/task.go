package task

import (
	"context"
	"fmt"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
)

func RunDev(store *logic.Store) error {
	ctx := context.Background()

	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{
				Name:     "user_id",
				Value:    1,
				Operator: "=",
			},
		},
	}

	sorting := repo.Sorting{
		Field: "amount",
		Order: "asc",
	}

	pagination := repo.Pagination{
		PerPage: 1,
		Page:    1,
	}

	queryOptions := repo.QueryOptions{
		Filters:    filters,
		Sorting:    sorting,
		Pagination: pagination,
	}

	es, err := store.FindExpenses(ctx, queryOptions)
	if err != nil {
		return err
	}

	c, err := store.CountExpenses(ctx, queryOptions.Filters)
	if err != nil {
		return err
	}

	for _, e := range es {
		fmt.Println(e)
	}
	fmt.Println("------------------")
	fmt.Println(c)

	return nil
}

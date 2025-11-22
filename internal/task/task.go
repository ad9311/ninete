package task

import (
	"context"
	"fmt"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
)

func RunDev(store *logic.Store) error {
	queryOptions := repo.QueryOptions{
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{
					Name:     "user_id",
					Value:    1,
					Operator: "=",
				},
			},
		},
		Sorting: repo.Sorting{
			Field: "amount",
			Order: "asc",
		},
		Pagination: repo.Pagination{
			PerPage: 2,
			Page:    2,
		},
	}

	es, err := store.FindExpenses(context.Background(), queryOptions)
	if err != nil {
		return err
	}

	for _, e := range es {
		fmt.Println(e)
	}

	return nil
}

package task

import (
	"context"
	"fmt"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
)

func RunDev(store *logic.Store) error {
	// queryOptions := repo.QueryOptions{
	// 	Filters: repo.Filters{
	// 		FilterFields: []repo.FilterField{
	// 			{
	// 				Name:     "user_id",
	// 				Value:    1,
	// 				Operator: "=",
	// 			},
	// 		},
	// 	},
	// 	Sorting: repo.Sorting{
	// 		Field: "amount",
	// 		Order: "asc",
	// 	},
	// 	Pagination: repo.Pagination{
	// 		PerPage: 2,
	// 		Page:    2,
	// 	},
	// }

	c, err := store.CountExpenses(context.Background(), repo.Filters{
		FilterFields: []repo.FilterField{
			{
				Name:     "user_id",
				Value:    1,
				Operator: "=",
			},
		},
	})
	if err != nil {
		return err
	}

	fmt.Println(c)

	// for _, e := range es {
	// 	fmt.Println(e)
	// }

	return nil
}

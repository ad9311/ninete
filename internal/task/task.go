package task

import (
	"fmt"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
)

func RunDev(_ *logic.Store) error {
	queryOptions := repo.QueryOptions{
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{
					Name:     "id",
					Value:    1,
					Operator: "=",
				},
				{
					Name:     "username",
					Value:    "Andres",
					Operator: "=",
				},
			},
			Connector: "AND",
		},
		Sorting: repo.Sorting{
			Field: "amount",
			Order: "desc",
		},
		Pagination: repo.Pagination{
			PerPage: 0,
			Page:    0,
		},
	}

	subQuery, err := queryOptions.Build()
	if err != nil {
		return err
	}

	fmt.Println(subQuery)

	return nil
}

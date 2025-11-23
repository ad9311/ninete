package repo_test

import (
	"fmt"
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestBuildQueryOptions(t *testing.T) {
	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{
				Name:     "id",
				Value:    1,
				Operator: "=",
			},
			{
				Name:     "sign_in_count",
				Value:    100,
				Operator: ">",
			},
		},
		Connector: "and",
	}

	sorting := repo.Sorting{
		Field: "sign_in_count",
		Order: "desc",
	}

	pagination := repo.Pagination{
		PerPage: 5,
		Page:    1,
	}

	opts := repo.QueryOptions{
		Filters:    filters,
		Sorting:    sorting,
		Pagination: pagination,
	}

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_build_correctly",
			func(t *testing.T) {
				subQuery, err := opts.Build()
				require.NoError(t, err)
				offset := (opts.Pagination.Page - 1) * opts.Pagination.PerPage
				expected := fmt.Sprintf(
					"WHERE \"%s\" %s ? AND \"%s\" %s ? ORDER BY \"%s\" %s LIMIT %d OFFSET %d",
					opts.Filters.FilterFields[0].Name,
					opts.Filters.FilterFields[0].Operator,
					opts.Filters.FilterFields[1].Name,
					opts.Filters.FilterFields[1].Operator,
					opts.Sorting.Field,
					opts.Sorting.Order,
					opts.Pagination.PerPage,
					offset,
				)
				require.Equal(t, expected, subQuery)
			},
		},
		{
			"should_fail_filters",
			func(t *testing.T) {
				failedOpts := opts

				failedOpts.Filters.Connector = ""
				_, err := failedOpts.Build()
				require.Error(t, err)
				require.ErrorIs(t, err, repo.ErrInvalidConnector)

				failedOpts.Filters = repo.Filters{
					FilterFields: []repo.FilterField{
						{
							Name:     "name",
							Value:    "value",
							Operator: "?",
						},
					},
				}
				_, err = failedOpts.Build()
				require.Error(t, err)
				require.ErrorIs(t, err, repo.ErrInvalidOperator)
			},
		},
		{
			"should_fail_sorting",
			func(t *testing.T) {
				failedOpts := repo.QueryOptions{
					Sorting: repo.Sorting{
						Field: "name",
						Order: "",
					},
				}
				_, err := failedOpts.Build()
				require.Error(t, err)
				require.ErrorIs(t, err, repo.ErrInvalidSortOrder)
			},
		},
		{
			"should_fail_pagination",
			func(t *testing.T) {
				failedOpts := repo.QueryOptions{}
				failedOpts.Pagination.Page = 1
				failedOpts.Pagination.PerPage = 0
				_, err := failedOpts.Build()
				require.Error(t, err)
				require.ErrorIs(t, err, repo.ErrInvalidPagination)
			},
		},
		{
			"should_not_fail_empty_options",
			func(t *testing.T) {
				opts := repo.QueryOptions{}
				subQuery, err := opts.Build()
				require.NoError(t, err)
				require.Empty(t, subQuery)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

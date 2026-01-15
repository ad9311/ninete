package task

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/seed"
)

type Config struct {
	App   *prog.App
	SQLDB *sql.DB
	Store *logic.Store
}

func New() (*Config, error) {
	app, err := prog.Load()
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.Open()
	if err != nil {
		return nil, err
	}

	queries := repo.New(app, sqlDB)

	store, err := logic.New(app, queries)
	if err != nil {
		return nil, err
	}

	tc := Config{
		App:   app,
		SQLDB: sqlDB,
		Store: store,
	}

	return &tc, nil
}

func (c *Config) RunTestCode() error {
	return nil
}

func (c *Config) CreateCategories() error {
	sc := seed.Config{
		App:   c.App,
		SQLDB: c.SQLDB,
	}

	if err := sc.SeedCategories(); err != nil {
		return err
	}

	return nil
}

func (c *Config) CreateExpensesFromRecurrent(ctx context.Context) error {
	const batchSize = 100
	created := 0
	nowUnix := time.Now().Unix()
	c.App.Logger.Log("create_expenses_from_recurrent started")
	defer func() {
		c.App.Logger.Logf("create_expenses_from_recurrent finished: created=%d", created)
	}()

	for {
		recurrentExpenses, err := c.Store.FindDueRecurrentExpenses(
			ctx,
			nowUnix,
			repo.Sorting{
				Field: "id",
				Order: "ASC",
			},
			repo.Pagination{
				PerPage: batchSize,
				Page:    1,
			},
		)
		if err != nil {
			return err
		}
		if len(recurrentExpenses) == 0 {
			return nil
		}

		for _, recurrent := range recurrentExpenses {
			_, err := c.Store.CreateExpenseFromPeriod(ctx, recurrent)
			if err != nil {
				if errors.Is(err, logic.ErrRecordAlreadyExist) {
					continue
				}

				return err
			}

			created++
		}

		if len(recurrentExpenses) < batchSize {
			return nil
		}
	}
}

func (c *Config) DeleteExpiredRefreshTokens(ctx context.Context) error {
	c.App.Logger.Log("Deleting expired refresh tokens started")
	defer func() {
		c.App.Logger.Log("Deleting expired refresh tokens finished")
	}()

	deleted, err := c.Store.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		return err
	}

	c.App.Logger.Logf("Deleted = %d", deleted)

	return nil
}

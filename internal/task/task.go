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

func (c *Config) CreateExpensesFromRecurrent(ctx context.Context) (int, error) {
	const batchSize = 100
	created := 0
	offset := 0
	nowUnix := time.Now().Unix()

	for {
		recurrentExpenses, err := c.Store.FindDueRecurrentExpenses(ctx, nowUnix, batchSize, offset)
		if err != nil {
			return created, err
		}
		if len(recurrentExpenses) == 0 {
			return created, nil
		}

		for _, recurrent := range recurrentExpenses {
			_, err := c.Store.CreateExpenseFromPeriod(ctx, recurrent)
			if err != nil {
				if errors.Is(err, logic.ErrRecordAlreadyExist) {
					continue
				}

				return created, err
			}

			created++
		}

		if len(recurrentExpenses) < batchSize {
			return created, nil
		}

		offset += batchSize
	}
}

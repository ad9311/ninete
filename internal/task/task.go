package task

import (
	"context"
	"database/sql"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/seed"
)

type Config struct {
	App     *prog.App
	SQLDB   *sql.DB
	Store   *logic.Store
	Context context.Context
}

func (c *Config) RunTestCode() error {
	return nil
}

func (c *Config) CreateCategories() error {
	sc := seed.Config{
		App:     c.App,
		SQLDB:   c.SQLDB,
		Context: c.Context,
	}

	if err := sc.SeedCategories(); err != nil {
		return err
	}

	return nil
}

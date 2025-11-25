package task

import (
	"database/sql"

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

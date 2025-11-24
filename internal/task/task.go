package task

import (
	"database/sql"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
)

type Config struct {
	App   *prog.App
	SQLDB *sql.DB
	Store *logic.Store
}

func (c *Config) RunDev() error {
	return nil
}

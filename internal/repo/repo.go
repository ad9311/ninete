package repo

import (
	"database/sql"

	"github.com/ad9311/ninete/internal/prog"
)

type Queries struct {
	app *prog.App
	db  *sql.DB
}

func New(app *prog.App, db *sql.DB) Queries {
	return Queries{
		app: app,
		db:  db,
	}
}

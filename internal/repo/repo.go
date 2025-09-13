package repo

import (
	"database/sql"
	"time"

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

func (q *Queries) wrapQuery(query string, queryFunc func()) {
	start := time.Now()
	defer func() {
		q.app.Logger.Query(query, time.Since(start))
	}()

	queryFunc()
}

package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/ad9311/ninete/internal/prog"
)

type QueryInt interface{}

type Queries struct {
	app *prog.App
	db  *sql.DB
}

type TxQueries struct {
	app *prog.App
	tx  *sql.Tx
}

func New(app *prog.App, db *sql.DB) Queries {
	return Queries{
		app: app,
		db:  db,
	}
}

func (q *Queries) WithTx(ctx context.Context, fn func(*TxQueries) error) error {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	tq := &TxQueries{app: q.app, tx: tx}
	if err := fn(tq); err != nil {
		_ = tx.Rollback()

		return err
	}

	return tx.Commit()
}

func wrapQuery(logger *prog.Logger, query string, queryFunc func() error) error {
	if !logger.EnableQuery {
		err := queryFunc()

		return err
	}

	start := time.Now()
	defer func() {
		logger.Query(query, time.Since(start))
	}()

	return queryFunc()
}

func (q *Queries) wrapQuery(query string, queryFunc func() error) error {
	return wrapQuery(q.app.Logger, query, queryFunc)
}

func (q *TxQueries) wrapQuery(query string, queryFunc func() error) error {
	return wrapQuery(q.app.Logger, query, queryFunc)
}

func newUpdatedAt() int64 {
	return time.Now().Unix()
}

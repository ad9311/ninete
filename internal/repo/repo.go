package repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ad9311/ninete/internal/prog"
)

type Queries struct {
	app *prog.App
	db  *sql.DB
}

type DBConnStats struct {
	MaxOpenConnections int `json:"maxOpenConnections"`
	IdleConnections    int `json:"idleConnections"`
	InUseConnections   int `json:"inUseConnections"`
}

func New(app *prog.App, db *sql.DB) Queries {
	return Queries{
		app: app,
		db:  db,
	}
}

func (q *Queries) CheckDBStatus() (DBConnStats, error) {
	var stats DBConnStats

	if err := q.db.Ping(); err != nil {
		return stats, fmt.Errorf("failed to ping database: %w", err)
	}

	stats = DBConnStats{
		MaxOpenConnections: q.db.Stats().MaxOpenConnections,
		IdleConnections:    q.db.Stats().Idle,
		InUseConnections:   q.db.Stats().InUse,
	}

	return stats, nil
}

func (q *Queries) wrapQuery(query string, queryFunc func()) {
	if !q.app.Logger.EnableQuery {
		queryFunc()

		return
	}

	start := time.Now()
	defer func() {
		q.app.Logger.Query(query, time.Since(start))
	}()

	queryFunc()
}

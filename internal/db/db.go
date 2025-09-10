package db

import (
	"database/sql"

	"github.com/ad9311/ninete/internal/conf"
	_ "github.com/mattn/go-sqlite3" // Database driver
)

func Open(dbc conf.DBConf) (*sql.DB, error) {
	var sqlDB *sql.DB

	sqlDB, err := sql.Open("sqlite3", "file:"+dbc.URL)
	if err != nil {
		return sqlDB, err // ERROR?
	}

	sqlDB.SetMaxOpenConns(dbc.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dbc.MaxIdleConns)

	return sqlDB, nil
}

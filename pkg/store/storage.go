package store

import (
	"database/sql"
	"github.com/jhachmer/gotocollection/pkg/config"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func NewSQLiteStorage(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:goto.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

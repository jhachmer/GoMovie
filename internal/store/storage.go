package store

import (
	"database/sql"

	"github.com/jhachmer/gotocollection/internal/config"
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

func SetupDatabase() (*SQLiteStorage, error) {
	db, err := NewSQLiteStorage(config.Envs)
	if err != nil {
		return nil, err
	}
	dbStore := NewStore(db)
	dbStore.TestDBConnection()
	err = dbStore.InitDatabaseTables()
	if err != nil {
		return nil, err
	}
	return dbStore, nil
}

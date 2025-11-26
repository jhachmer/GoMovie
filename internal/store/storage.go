package store

import (
	"database/sql"
	"fmt"

	"github.com/jhachmer/gomovie/internal/config"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func NewStorage(cfg config.Config) (Store, error) {
	db, err := sql.Open(cfg.DbType, cfg.DbConfig.ConnectionString())
	if err != nil {
		return nil, err
	}
	switch cfg.DbType {
	case "sqlite3":
		return NewSQLiteStore(db), nil
	case "postgres":
		return NewPostgresStore(db), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DbType)
	}
}

func SetupDatabase(cfg config.Config) (Store, error) {
	store, err := NewStorage(cfg)
	if err != nil {
		return nil, err
	}

	if err := store.TestDBConnection(); err != nil {
		return nil, err
	}

	if err := store.InitDatabaseTables(); err != nil {
		return nil, err
	}

	if err := store.CreateAdminAccount(cfg); err != nil {
		return nil, err
	}

	return store, nil
}

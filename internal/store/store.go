package store

import (
	"database/sql"

	"github.com/jhachmer/gomovie/internal/config"
	"github.com/jhachmer/gomovie/internal/types"
)

type Store interface {
	TestDBConnection() error
	InitDatabaseTables() error
	CreateAdminAccount(config config.Config) error
	Close() error
	UserStore
	MediaStore
	EntryStore
	StatsStore
}

type UserStore interface {
	CreateUser(username, password string) error
	CheckCredentials(username, password string) (bool, error)
	AdminLoginQuery(username string) (string, error)
	GetUsers() (*sql.Rows, error)
	ToggleUserActive(userID, status int) error
}

type MediaStore interface {
	CreateMovie(*types.Movie) (*types.Movie, error)
	UpdateMovie(*types.Movie) (*types.Movie, error)
	GetMovieByID(string) (*types.Movie, error)
	GetAllMovies() ([]*types.MovieInfoData, error)

	CreateSeries(*types.Series) (*types.Series, error)
	UpdateSeries(*types.Series) (*types.Series, error)
	//GetSeriesByID(string) (*types.Series, error)
	//GetAllSeries() ([]*types.SeriesInfoData, error)

	DeleteMedia(string) error

	SearchMovie(types.SearchParams) ([]*types.MovieInfoData, error)
}

type EntryStore interface {
	CreateEntry(entry *types.Entry, movie *types.Movie) (*types.Entry, error)
	GetEntries(userID string) ([]*types.Entry, error)
	UpdateEntry(entryID, field, newValue string, watched bool) (*types.Entry, error)
	DeleteEntry(entryID string) error
}

type StatsStore interface {
	GetWatchCounts() (*types.WatchStats, error)
}

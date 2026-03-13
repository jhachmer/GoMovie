package store

import (
	"database/sql"

	"github.com/jhachmer/gomovie/internal/api"
	"github.com/jhachmer/gomovie/internal/config"
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
	CreateMovie(*api.Movie) (*api.Movie, error)
	UpdateMovie(*api.Movie) (*api.Movie, error)
	GetMovieByID(string) (*api.Movie, error)
	GetAllMovies() ([]*api.MovieInfoData, error)

	CreateSeries(*api.Series) (*api.Series, error)
	UpdateSeries(*api.Series) (*api.Series, error)
	//GetSeriesByID(string) (*types.Series, error)
	//GetAllSeries() ([]*types.SeriesInfoData, error)

	DeleteMedia(string) error

	SearchMovie(api.SearchParams) ([]*api.MovieInfoData, error)
}

type EntryStore interface {
	CreateEntry(entry *api.Entry, movie *api.Movie) (*api.Entry, error)
	GetEntries(userID string) ([]*api.Entry, error)
	UpdateEntry(entryID, field, newValue string, watched bool) (*api.Entry, error)
	DeleteEntry(entryID string) error
}

type StatsStore interface {
	GetWatchCounts() (*api.WatchStats, error)
}

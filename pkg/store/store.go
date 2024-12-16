package store

import (
	"database/sql"
	"github.com/jhachmer/gotocollection/pkg/media"
)

type Storage struct {
	db *sql.DB
}
type Store interface {
	InitDatabase() error
	CreateEntry(entry *media.Entry) (*media.Entry, error)
	GetEntries(id string) ([]media.Entry, error)
	CreateMovie(mov media.Movie) (*media.Movie, error)
	GetMovie(id string) (*media.Movie, error)
}

func NewStore(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) InitDatabase() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS movies (
		id VARCHAR(9) NOT NULL,
		title VARCHAR(255) NOT NULL,
		year VARCHAR(255) NOT NULL,
		genre VARCHAR(255) NOT NULL,
		actors VARCHAR(500) NOT NULL,
    	director VARCHAR(500) NOT NULL,
    	runtime VARCHAR(500) NOT NULL,
    	rated VARCHAR(255) NOT NULL,
    	released VARCHAR(500) NOT NULL,
    	plot TEXT NOT NULL,
    	poster VARCHAR(500) NOT NULL,

		PRIMARY KEY (id)
	);`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS ratings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id VARCHAR(9) NOT NULL,
		source VARCHAR(255) NOT NULL,
		value VARCHAR(50) NOT NULL,

		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
		);`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL,
		watched INTEGER DEFAULT 0,
		comment TEXT,
		movie_id VARCHAR(9) NOT NULL,

		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE SET NULL
		);`)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) CreateEntry(e *media.Entry) (*media.Entry, error) {
	_, err := s.CreateMovie(e.Movie)
	if err != nil {
		return nil, err
	}
	var watchedInt = 0
	if e.Watched {
		watchedInt = 1
	}
	res, err := s.db.Exec(`INSERT INTO entries (name, watched, comment, movie_id)
		VALUES (?, ?, ?, ?)`,
		e.Name, watchedInt, e.Comment, e.Movie.ImdbID)
	if err != nil {
		return nil, err
	}
	e.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (s *Storage) CreateMovie(m *media.Movie) (*media.Movie, error) {
	_, err := s.db.Exec(`INSERT INTO movies (id, title, year, genre, actors, director, runtime, rated, released, plot, poster)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ImdbID, m.Title, m.Year, m.Genre, m.Actors, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster)
	if err != nil {
		return nil, err
	}
	for _, rating := range m.Ratings {
		_, err = s.db.Exec(`INSERT INTO ratings (movie_id, source, value)
		VALUES (?, ?, ?)`,
			m.ImdbID, rating.Source, rating.Value)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jhachmer/gotocollection/internal/media"
)

type Storage struct {
	db *sql.DB
}
type Store interface {
	InitDatabase() error
	CreateEntry(entry *media.InfoPage) (*media.InfoPage, error)
	GetEntries(id string) ([]media.InfoPage, error)
	CreateMovie(mov media.Movie) (*media.Movie, error)
	GetMovie(id string) (*media.Movie, error)
}

func NewStore(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) TestDBConnection() {
	err := s.db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to DB...")
}

func (s *Storage) InitDatabase() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS movies (
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

func (s *Storage) CreateEntry(e *media.Entry, mov *media.Movie) (*media.Entry, error) {
	var exists bool
	row := s.db.QueryRow(`SELECT EXISTS(SELECT movies.title FROM movies WHERE movies.id = ?)`, mov.ImdbID)
	if err := row.Scan(&exists); err != nil {
		log.Println("movie exists:", exists)
		return nil, err
	} else if !exists {
		_, err := s.CreateMovie(mov)
		if err != nil {
			return nil, err
		}
	}
	var watchedInt = 0
	if e.Watched {
		watchedInt = 1
	}
	res, err := s.db.Exec(`INSERT INTO entries (name, watched, comment, movie_id)
		VALUES (?, ?, ?, ?)`,
		e.Name, watchedInt, e.Comment, mov.ImdbID)
	if err != nil {
		return nil, err
	}
	e.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (s *Storage) GetEntries(id string) ([]*media.Entry, error) {
	rows, err := s.db.Query(`SELECT id, name, watched, comment FROM entries WHERE movie_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*media.Entry

	for rows.Next() {
		var entry media.Entry
		if err := rows.Scan(&entry.ID, &entry.Name, &entry.Watched, &entry.Comment); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
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

func (s *Storage) GetMovie(id string) (*media.Movie, error) {
	var mov media.Movie
	if err := s.db.QueryRow(`SELECT * FROM movies WHERE id = ?`, id).Scan(&mov.ImdbID, &mov.Title, &mov.Year, &mov.Genre, &mov.Actors, &mov.Director, &mov.Runtime, &mov.Rated, &mov.Released, &mov.Plot, &mov.Poster); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("could not find movie with id %s", id)
		}
		return nil, err
	}
	ratings, err := s.getRatings(id)
	if err != nil {
		return nil, err
	}
	mov.Ratings = ratings
	return &mov, nil
}

func (s *Storage) getRatings(id string) ([]media.Rating, error) {
	rows, err := s.db.Query(`SELECT source, value FROM ratings WHERE movie_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []media.Rating

	for rows.Next() {
		var rating media.Rating
		if err := rows.Scan(&rating.Source, &rating.Value); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ratings, nil
}

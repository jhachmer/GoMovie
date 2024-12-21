package store

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jhachmer/gotocollection/internal/types"
	"log"
)

type Storage struct {
	db *sql.DB
}
type Store interface {
	InitDatabase() error
	CreateEntry(*types.Entry, *types.Movie) (*types.Entry, error)
	GetEntries(string) ([]*types.Entry, error)
	CreateMovie(*types.Movie) (*types.Movie, error)
	GetMovie(id string) (*types.Movie, error)
	GetAllMovies() ([]*types.Movie, error)
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

func (s *Storage) CreateEntry(e *types.Entry, mov *types.Movie) (*types.Entry, error) {
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

func (s *Storage) GetEntries(id string) ([]*types.Entry, error) {
	rows, err := s.db.Query(`SELECT id, name, watched, comment FROM entries WHERE movie_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*types.Entry

	for rows.Next() {
		var entry types.Entry
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

func (s *Storage) CreateMovie(m *types.Movie) (*types.Movie, error) {
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

func (s *Storage) GetMovie(id string) (*types.Movie, error) {
	var mov types.Movie
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

func (s *Storage) GetAllMovies() ([]*types.Movie, error) {
	rows, err := s.db.Query(`SELECT id, title, year FROM movies ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*types.Movie

	for rows.Next() {
		var movie types.Movie
		if err := rows.Scan(&movie.ImdbID, &movie.Title, &movie.Year); err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}

func (s *Storage) getRatings(id string) ([]types.Rating, error) {
	rows, err := s.db.Query(`SELECT source, value FROM ratings WHERE movie_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []types.Rating

	for rows.Next() {
		var rating types.Rating
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

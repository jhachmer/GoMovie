package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jhachmer/gotocollection/internal/types"
	"github.com/jhachmer/gotocollection/internal/util"
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
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS genres (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE
	);`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS actors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE
	);`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS movies_genres (
    movie_id VARCHAR(9) NOT NULL,
    genre_id INTEGER NOT NULL,
    PRIMARY KEY (movie_id, genre_id),
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES genres(id) ON DELETE CASCADE
);`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS movies_actors (
    movie_id VARCHAR(9) NOT NULL,
    actor_id INTEGER NOT NULL,
    PRIMARY KEY (movie_id, actor_id),
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_id) REFERENCES actors(id) ON DELETE CASCADE
);
`)
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
	_, err := s.db.Exec(`INSERT INTO movies (id, title, year, director, runtime, rated, released, plot, poster)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ImdbID, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster)
	if err != nil {
		return nil, err
	}
	err = s.createRatings(m, err)
	if err != nil {
		return nil, err
	}
	err = s.createGenres(m)
	if err != nil {
		return nil, err
	}
	err = s.createActors(m)
	return m, nil
}

func (s *Storage) GetMovie(id string) (*types.Movie, error) {
	var mov types.Movie
	if err := s.db.QueryRow(`SELECT * FROM movies WHERE id = ?`, id).Scan(&mov.ImdbID, &mov.Title, &mov.Year, &mov.Director, &mov.Runtime, &mov.Rated, &mov.Released, &mov.Plot, &mov.Poster); err != nil {
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
	actors, err := s.getActors(id)
	if err != nil {
		return nil, err
	}
	mov.Actors = actors
	genres, err := s.getGenres(id)
	if err != nil {
		return nil, err
	}
	mov.Genre = genres
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

func (s *Storage) createRatings(m *types.Movie, err error) error {
	for _, rating := range m.Ratings {
		_, err = s.db.Exec(`INSERT INTO ratings (movie_id, source, value)
		VALUES (?, ?, ?)`,
			m.ImdbID, rating.Source, rating.Value)
		if err != nil {
			return err
		}
	}
	return nil
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

func (s *Storage) createGenres(m *types.Movie) error {
	genres := util.SplitIMDBString(m.Genre)
	for _, genre := range genres {
		var genreID int64
		err := s.db.QueryRow("SELECT id FROM genres WHERE name = ?", genre).Scan(&genreID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := s.db.Exec("INSERT OR IGNORE INTO genres (name) VALUES (?)", genre)
			if err != nil {
				return err
			}
			genreID, _ = res.LastInsertId()
		}
		_, err = s.db.Exec(`INSERT OR IGNORE INTO movies_genres (movie_id, genre_id) VALUES (?, ?);`, m.ImdbID, genreID)
	}
	return nil
}

func (s *Storage) getGenres(id string) (string, error) {
	rows, err := s.db.Query(`SELECT g.name 
		FROM genres g
		JOIN movies_genres mg ON g.id = mg.genre_id
		WHERE mg.movie_id = ?;
		`, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var genres []string

	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			return "", err
		}
		genres = append(genres, genre)
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return util.JoinIMDBStrings(genres), nil
}

func (s *Storage) createActors(m *types.Movie) error {
	actors := util.SplitIMDBString(m.Actors)
	for _, actor := range actors {
		var actorID int64
		err := s.db.QueryRow("SELECT id FROM actors WHERE name = ?", actor).Scan(&actorID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := s.db.Exec("INSERT OR IGNORE INTO actors (name) VALUES (?)", actor)
			if err != nil {
				return err
			}
			actorID, _ = res.LastInsertId()
		}
		_, err = s.db.Exec(`INSERT OR IGNORE INTO movies_actors (movie_id, actor_id) VALUES (?, ?);`, m.ImdbID, actorID)
	}
	return nil
}

func (s *Storage) getActors(id string) (string, error) {
	rows, err := s.db.Query(`SELECT a.name 
		FROM actors a
		JOIN movies_actors ma ON a.id = ma.actor_id
		WHERE ma.movie_id = ?;
		`, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var actors []string

	for rows.Next() {
		var actor string
		if err := rows.Scan(&actor); err != nil {
			return "", err
		}
		actors = append(actors, actor)
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return util.JoinIMDBStrings(actors), nil
}

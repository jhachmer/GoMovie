package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jhachmer/gotocollection/internal/types"
	"github.com/jhachmer/gotocollection/internal/util"
	"golang.org/x/crypto/bcrypt"
)

type SQLiteStorage struct {
	db *sql.DB
}

// TODO: add UpdateMovie
type Store interface {
	InitDatabaseTables() error
	CheckCredentials(string, string) (bool, error)
	CreateEntry(*types.Entry, *types.Movie) (*types.Entry, error)
	GetEntries(string) ([]*types.Entry, error)
	UpdateEntry(string, string, string, bool) (*types.Entry, error)
	DeleteEntry(string) error
	CreateMovie(*types.Movie) (*types.Movie, error)
	GetMovieByID(string) (*types.Movie, error)
	GetAllMovies() ([]*types.MovieOverviewData, error)
	SearchMovie(SearchParams) ([]*types.MovieOverviewData, error)
}

func NewStore(db *sql.DB) *SQLiteStorage {
	return &SQLiteStorage{
		db: db,
	}
}

func (s *SQLiteStorage) Close() {
	s.db.Close()
}

func (s *SQLiteStorage) TestDBConnection() {
	err := s.db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to DB...")
}

func (s *SQLiteStorage) InitDatabaseTables() error {
	_, err := s.db.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS useraccounts (
    	UserID INTEGER PRIMARY KEY AUTOINCREMENT,
    	Username TEXT NOT NULL UNIQUE,
    	PasswordHash TEXT NOT NULL);
		`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec( /*sql*/ `
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

		PRIMARY KEY (id));
		`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS ratings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id VARCHAR(9) NOT NULL,
		source VARCHAR(255) NOT NULL,
		value VARCHAR(50) NOT NULL,

		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE);
		`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL,
		watched INTEGER DEFAULT 0,
		comment TEXT,
		movie_id VARCHAR(9) NOT NULL UNIQUE,
		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE SET NULL);
		`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS genres (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL UNIQUE);
		`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS actors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL UNIQUE);
		`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS movies_genres (
		movie_id VARCHAR(9) NOT NULL,
		genre_id INTEGER NOT NULL,
		PRIMARY KEY (movie_id, genre_id),
		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
		FOREIGN KEY (genre_id) REFERENCES genres(id) ON DELETE CASCADE);
		`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS movies_actors (
		movie_id VARCHAR(9) NOT NULL,
		actor_id INTEGER NOT NULL,
		PRIMARY KEY (movie_id, actor_id),
		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
		FOREIGN KEY (actor_id) REFERENCES actors(id) ON DELETE CASCADE);
		`)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) CheckCredentials(username, password string) (bool, error) {
	var hashedPassword string

	err := s.db.QueryRow( /*sql*/ `
		SELECT PasswordHash
		FROM UserAccounts
		WHERE Username = ?
		`, username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (s *SQLiteStorage) CreateEntry(e *types.Entry, mov *types.Movie) (*types.Entry, error) {
	var exists bool
	row := s.db.QueryRow( /*sql*/ `
		SELECT EXISTS(SELECT movies.title
		FROM movies
		WHERE movies.id = ?);
		`, mov.ImdbID)
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
	res, err := s.db.Exec( /*sql*/ `
		INSERT INTO entries
		(name, watched, comment, movie_id)
		VALUES (?, ?, ?, ?);
		`, e.Name, watchedInt, e.Comment, mov.ImdbID)
	if err != nil {
		return nil, err
	}
	e.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *SQLiteStorage) UpdateEntry(movieId, name, comment string, watched bool) (*types.Entry, error) {
	var watchedInt = 0
	if watched {
		watchedInt = 1
	}
	res, err := s.db.Exec( /*sql*/ `
		UPDATE entries
		SET name = ?, comment = ?, watched = ?
		WHERE movie_id = ?
	`, name, comment, watchedInt, movieId)
	if err != nil {
		return nil, err
	}
	resID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	entry := types.Entry{
		ID:      resID,
		Name:    name,
		Comment: []byte(comment),
		Watched: watched,
	}
	return &entry, nil
}

func (s *SQLiteStorage) DeleteEntry(imdbId string) error {
	_, err := s.db.Exec( /*sql*/ `
		DELETE FROM entries
		WHERE movie_id = ?
	`, imdbId)
	if err != nil {
		return fmt.Errorf("error deleting entry for movie: %s\n%w", imdbId, err)
	}
	return nil
}

func (s *SQLiteStorage) GetEntries(id string) ([]*types.Entry, error) {
	rows, err := s.db.Query( /*sql*/ `
		SELECT id, name, watched, comment
		FROM entries
		WHERE movie_id = ?;
		`, id)
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

func (s *SQLiteStorage) CreateMovie(m *types.Movie) (*types.Movie, error) {
	_, err := s.db.Exec( /*sql*/ `
		INSERT INTO movies (id, title, year, director, runtime, rated, released, plot, poster)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
		`, m.ImdbID, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster)
	if err != nil {
		return nil, err
	}
	err = s.createRatings(m)
	if err != nil {
		return nil, err
	}
	err = s.createGenres(m)
	if err != nil {
		return nil, err
	}
	err = s.createActors(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *SQLiteStorage) GetMovieByID(movieID string) (*types.Movie, error) {
	var movie types.Movie

	err := s.db.QueryRow( /*sql*/ `
        SELECT
            id, title, year, rated, released, runtime, plot, poster, director
        FROM movies
        WHERE id = ?`, movieID).Scan(
		&movie.ImdbID, &movie.Title, &movie.Year, &movie.Rated,
		&movie.Released, &movie.Runtime, &movie.Plot, &movie.Poster, &movie.Director)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query( /*sql*/ `
        SELECT g.name
        FROM genres g
        INNER JOIN movies_genres mg ON g.id = mg.genre_id
        WHERE mg.movie_id = ?`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []string
	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	movie.Genre = strings.Join(genres, ", ")

	rows, err = s.db.Query( /*sql*/ `
        SELECT a.name
        FROM actors a
        INNER JOIN movies_actors ma ON a.id = ma.actor_id
        WHERE ma.movie_id = ?`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actors []string
	for rows.Next() {
		var actor string
		if err := rows.Scan(&actor); err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}
	movie.Actors = strings.Join(actors, ", ")

	rows, err = s.db.Query( /*sql*/ `
        SELECT source, value
        FROM ratings
        WHERE movie_id = ?`, movieID)
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
	movie.Ratings = ratings

	return &movie, nil
}

func (s *SQLiteStorage) GetAllMovies() ([]*types.MovieOverviewData, error) {
	var movies []*types.MovieOverviewData
	rows, err := s.db.Query( /*sql*/ `
        SELECT id
        FROM movies
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movieIDs []string
	for rows.Next() {
		var movieID string
		if err := rows.Scan(&movieID); err != nil {
			return nil, err
		}
		movieIDs = append(movieIDs, movieID)
	}
	for _, movieID := range movieIDs {
		movie, err := s.GetMovieByID(movieID)
		if err != nil {
			return nil, err
		}
		entries, err := s.GetEntries(movieID)
		if err != nil {
			return nil, err
		}
		data := types.MovieOverviewData{Movie: movie, Entry: entries}
		movies = append(movies, &data)
	}
	return movies, nil
}

type SearchParams struct {
	Genres []string
	Actors []string
	Years  YearSearch
}

type YearSearch struct {
	StartYear string
	EndYear   string
}

func (s *Storage) SearchMovie(params SearchParams) ([]*types.MovieOverviewData, error) {
	filters := []string{}
	args := []any{}

	if len(params.Genres) > 0 {
		genrePlaceholders := make([]string, len(params.Genres))
		for i, genre := range params.Genres {
			genrePlaceholders[i] = "?"
			args = append(args, "%"+genre+"%")
		}
		filters = append(filters, "g.name LIKE "+strings.Join(genrePlaceholders, " OR g.name LIKE "))
	}

	if len(params.Actors) > 0 {
		actorPlaceholders := make([]string, len(params.Actors))
		for i, actor := range params.Actors {
			actorPlaceholders[i] = "?"
			args = append(args, "%"+actor+"%")
		}
		filters = append(filters, "a.name LIKE "+strings.Join(actorPlaceholders, " OR a.name LIKE "))
	}

	if params.Years.StartYear != "" && params.Years.EndYear != "" {
		filters = append(filters, "m.year BETWEEN ? AND ?")
		args = append(args, params.Years.StartYear, params.Years.EndYear)
	}

	query := /*sql*/ `
		SELECT DISTINCT m.id
		FROM movies m
		LEFT JOIN movies_genres mg ON m.id = mg.movie_id
		LEFT JOIN genres g ON mg.genre_id = g.id
		LEFT JOIN movies_actors ma ON m.id = ma.movie_id
		LEFT JOIN actors a ON ma.actor_id = a.id
		`

	if len(filters) > 0 {
		query += "WHERE " + strings.Join(filters, " AND ")
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	results := []*types.MovieOverviewData{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		mov, err := s.GetMovieByID(id)
		if err != nil {
			return nil, err
		}
		entry, err := s.GetEntries(id)
		if err != nil {
			return nil, err
		}
		results = append(results, &types.MovieOverviewData{Movie: mov, Entry: entry})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (s *SQLiteStorage) createRatings(m *types.Movie) error {
	for _, rating := range m.Ratings {
		_, err := s.db.Exec( /*sql*/ `
			INSERT INTO ratings (movie_id, source, value)
			VALUES (?, ?, ?);
			`, m.ImdbID, rating.Source, rating.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) getRatings(id string) ([]types.Rating, error) {
	rows, err := s.db.Query( /*sql*/ `
		SELECT source, value
		FROM ratings
		WHERE movie_id = ?;
		`, id)
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

func (s *SQLiteStorage) createGenres(m *types.Movie) error {
	genres := util.SplitIMDBString(m.Genre)
	for _, genre := range genres {
		var genreID int64
		err := s.db.QueryRow( /*sql*/ `
			SELECT id
			FROM genres
			WHERE name = ?;
			`, genre).Scan(&genreID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := s.db.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO genres (name)
       			VALUES (?);
       			`, genre)
			if err != nil {
				return err
			}
			genreID, _ = res.LastInsertId()
		}
		_, err = s.db.Exec( /*sql*/ `INSERT OR IGNORE
       		INTO movies_genres (movie_id, genre_id)
       		VALUES (?, ?);
       		`, m.ImdbID, genreID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) getGenres(id string) (string, error) {
	rows, err := s.db.Query( /*sql*/ `
		SELECT g.name
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

func (s *SQLiteStorage) createActors(m *types.Movie) error {
	actors := util.SplitIMDBString(m.Actors)
	for _, actor := range actors {
		var actorID int64
		err := s.db.QueryRow( /*sql*/ `
			SELECT id
			FROM actors
			WHERE name = ?;
			`, actor).Scan(&actorID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := s.db.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO actors (name)
       			VALUES (?);
       			`, actor)
			if err != nil {
				return err
			}
			actorID, _ = res.LastInsertId()
		}
		_, err = s.db.Exec( /*sql*/ `
			INSERT OR IGNORE
       		INTO movies_actors (movie_id, actor_id)
       		VALUES (?, ?);
			`, m.ImdbID, actorID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) getActors(id string) (string, error) {
	rows, err := s.db.Query( /*sql*/ `
		SELECT a.name
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

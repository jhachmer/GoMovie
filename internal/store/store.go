package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/jhachmer/gomovie/internal/auth"
	"github.com/jhachmer/gomovie/internal/types"
	"github.com/jhachmer/gomovie/internal/util"
	"golang.org/x/crypto/bcrypt"
)

type SQLiteStorage struct {
	DB *sql.DB
}

type Store interface {
	InitDatabaseTables() error
	CreateUser(string, string) error
	CheckCredentials(string, string) (bool, error)
	CreateEntry(*types.Entry, *types.Movie) (*types.Entry, error)
	GetEntries(string) ([]*types.Entry, error)
	UpdateEntry(string, string, string, bool) (*types.Entry, error)
	DeleteEntry(string) error
	CreateMovie(*types.Movie) (*types.Movie, error)
	UpdateMovie(*types.Movie) (*types.Movie, error)
	DeleteMovie(string) error
	GetMovieByID(string) (*types.Movie, error)
	GetAllMovies() ([]*types.MovieOverviewData, error)
	SearchMovie(types.SearchParams) ([]*types.MovieOverviewData, error)

	GetWatchCounts() (*types.WatchStats, error)

	AdminLoginQuery(string) (string, error)
	GetUsers() (*sql.Rows, error)
	ToggleUserActive(int, int) error
}

func NewStore(db *sql.DB) *SQLiteStorage {
	return &SQLiteStorage{
		DB: db,
	}
}

func (s *SQLiteStorage) Close() {
	s.DB.Close()
}

func (s *SQLiteStorage) TestDBConnection() {
	err := s.DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to DB...")
}

func (s *SQLiteStorage) CreateAdminAccount(name, pw string) error {
	hashedPW, err := auth.HashPassword(pw)
	if err != nil {
		log.Fatal("error creating admin account")
		return err
	}
	_, err = s.DB.Exec( /*sql*/ `
	INSERT OR IGNORE INTO useraccounts (Username, PasswordHash, Active, IsAdmin)
	VALUES (?, ?, ?, ?)
	`, name, hashedPW, 1, 1)
	if err != nil {
		return fmt.Errorf("error inserting admin acc %w", err)
	}
	return nil
}

func (s *SQLiteStorage) InitDatabaseTables() error {
	_, err := s.DB.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS useraccounts (
    	UserID INTEGER PRIMARY KEY AUTOINCREMENT,
    	Username TEXT NOT NULL UNIQUE,
    	PasswordHash TEXT NOT NULL,
		Active INTEGER DEFAULT 0,
		IsAdmin INTEGER DEFAULT 0);
		`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec( /*sql*/ `
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
	_, err = s.DB.Exec( /*sql*/ `
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
	_, err = s.DB.Exec( /*sql*/ `
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
	_, err = s.DB.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS genres (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL UNIQUE);
		`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec( /*sql*/ `
		CREATE TABLE IF NOT EXISTS actors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL UNIQUE);
		`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec( /*sql*/ `
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
	_, err = s.DB.Exec( /*sql*/ `
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
	var active bool

	err := s.DB.QueryRow( /*sql*/ `
		SELECT PasswordHash, Active
		FROM UserAccounts
		WHERE Username = ?
		`, username).Scan(&hashedPassword, &active)
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
	if !active {
		return false, fmt.Errorf("user account not activated")
	}
	return true, nil
}

func (s *SQLiteStorage) CreateEntry(e *types.Entry, mov *types.Movie) (*types.Entry, error) {
	var exists bool
	row := s.DB.QueryRow( /*sql*/ `
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
	res, err := s.DB.Exec( /*sql*/ `
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

func (s *SQLiteStorage) CreateEntryTx(tx *sql.Tx, e *types.Entry, mov *types.Movie) (*types.Entry, error) {
	var exists bool
	row := tx.QueryRow( /*sql*/ `
		SELECT EXISTS(SELECT movies.title
		FROM movies
		WHERE movies.id = ?);
		`, mov.ImdbID)
	if err := row.Scan(&exists); err != nil {
		log.Println("movie exists:", exists)
		return nil, err
	} else if !exists {
		_, err := s.CreateMovieTx(tx, mov)
		if err != nil {
			return nil, err
		}
	}
	var watchedInt = 0
	if e.Watched {
		watchedInt = 1
	}
	res, err := tx.Exec( /*sql*/ `
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
	res, err := s.DB.Exec( /*sql*/ `
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
	_, err := s.DB.Exec( /*sql*/ `
		DELETE FROM entries
		WHERE movie_id = ?
	`, imdbId)
	if err != nil {
		return fmt.Errorf("error deleting entry for movie: %s\n%w", imdbId, err)
	}
	return nil
}

func (s *SQLiteStorage) GetEntries(id string) ([]*types.Entry, error) {
	rows, err := s.DB.Query( /*sql*/ `
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
	_, err := s.DB.Exec( /*sql*/ `
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

func (s *SQLiteStorage) CreateMovieTx(tx *sql.Tx, m *types.Movie) (*types.Movie, error) {
	_, err := tx.Exec( /*sql*/ `
		INSERT INTO movies (id, title, year, director, runtime, rated, released, plot, poster)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
		`, m.ImdbID, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster)
	if err != nil {
		return nil, err
	}
	err = s.createRatingsTx(tx, m)
	if err != nil {
		return nil, err
	}
	err = s.createGenresTx(tx, m)
	if err != nil {
		return nil, err
	}
	err = s.createActorsTx(tx, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *SQLiteStorage) UpdateMovie(m *types.Movie) (*types.Movie, error) {
	_, err := s.DB.Exec( /*sql*/ `
	UPDATE movies
	SET title = ?, year = ?, director = ?, runtime = ?, rated = ?, released = ?, plot = ?, poster = ?
	WHERE id = ?;
	`, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster, m.ImdbID)
	if err != nil {
		return nil, err
	}
	err = s.updateRatings(m)
	if err != nil {
		return nil, err
	}
	err = s.updateGenres(m)
	if err != nil {
		return nil, err
	}
	err = s.updateActors(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *SQLiteStorage) DeleteMovie(imdbId string) error {
	_, err := s.DB.Exec( /*sql*/ `
	DELETE FROM movies WHERE id = ?;
	`, imdbId)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) updateRatings(m *types.Movie) error {
	for _, rating := range m.Ratings {
		_, err := s.DB.Exec( /*sql*/ `
			UPDATE ratings
			SET value = ?
			WHERE movie_id = ? AND source = ?;
			`, rating.Value, m.ImdbID, rating.Source)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) updateGenres(m *types.Movie) error {
	genres := util.SplitIMDBString(m.Genre)

	rows, err := s.DB.Query( /*sql*/ `
	SELECT g.name
	FROM genres g
	INNER JOIN movies_genres mg ON g.id = mg.genre_id
	WHERE mg.movie_id = ?`, m.ImdbID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var genresFromDB []string
	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			return err
		}
		genresFromDB = append(genresFromDB, genre)
	}

	for _, g := range genresFromDB {
		if !slices.Contains(genres, g) {
			_, err := s.DB.Exec( /*sql*/ `
			DELETE FROM movies_genres
			WHERE
			movie_id = ?
			AND
			genre_id = (SELECT id FROM genres WHERE name = ?);
			`, m.ImdbID, g)
			if err != nil {
				return err
			}
		}
	}

	for _, genre := range genres {
		var genreID int64
		err := s.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM genres
			WHERE name = ?;
			`, genre).Scan(&genreID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := s.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO genres (name)
       			VALUES (?);
       			`, genre)
			if err != nil {
				return err
			}
			genreID, _ = res.LastInsertId()
		}
		_, err = s.DB.Exec( /*sql*/ `
			INSERT OR IGNORE
       		INTO movies_genres (movie_id, genre_id)
       		VALUES (?, ?);
       		`, m.ImdbID, genreID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) updateActors(m *types.Movie) error {
	actors := util.SplitIMDBString(m.Actors)

	rows, err := s.DB.Query( /*sql*/ `
	SELECT a.name
	FROM actors a
	INNER JOIN movies_actors ma ON a.id = ma.actor_id
	WHERE ma.movie_id = ?`, m.ImdbID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var actorsFromDB []string
	for rows.Next() {
		var actor string
		if err := rows.Scan(&actor); err != nil {
			return err
		}
		actorsFromDB = append(actorsFromDB, actor)
	}

	for _, a := range actorsFromDB {
		if !slices.Contains(actors, a) {
			_, err := s.DB.Exec( /*sql*/ `
			DELETE FROM movies_actors
			WHERE
			movie_id = ?
			AND
			actor_id = (SELECT id FROM actors WHERE name = ?);
			`, m.ImdbID, a)
			if err != nil {
				return err
			}
		}
	}

	for _, actor := range actors {
		var actorID int64
		err := s.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM actors
			WHERE name = ?;
			`, actor).Scan(&actorID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := s.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO actors (name)
       			VALUES (?);
       			`, actor)
			if err != nil {
				return err
			}
			actorID, _ = res.LastInsertId()
		}
		_, err = s.DB.Exec( /*sql*/ `
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

func (s *SQLiteStorage) GetMovieByID(movieID string) (*types.Movie, error) {
	var movie types.Movie

	err := s.DB.QueryRow( /*sql*/ `
        SELECT
            id, title, year, rated, released, runtime, plot, poster, director
        FROM movies
        WHERE id = ?`, movieID).Scan(
		&movie.ImdbID, &movie.Title, &movie.Year, &movie.Rated,
		&movie.Released, &movie.Runtime, &movie.Plot, &movie.Poster, &movie.Director)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.Query( /*sql*/ `
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

	rows, err = s.DB.Query( /*sql*/ `
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

	rows, err = s.DB.Query( /*sql*/ `
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
	rows, err := s.DB.Query( /*sql*/ `
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

func (s *SQLiteStorage) SearchMovie(params types.SearchParams) ([]*types.MovieOverviewData, error) {
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

	rows, err := s.DB.Query(query, args...)
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
		_, err := s.DB.Exec( /*sql*/ `
			INSERT INTO ratings (movie_id, source, value)
			VALUES (?, ?, ?);
			`, m.ImdbID, rating.Source, rating.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) createRatingsTx(tx *sql.Tx, m *types.Movie) error {
	for _, rating := range m.Ratings {
		_, err := tx.Exec( /*sql*/ `
			INSERT INTO ratings (movie_id, source, value)
			VALUES (?, ?, ?);
			`, m.ImdbID, rating.Source, rating.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) createGenres(m *types.Movie) error {
	genres := util.SplitIMDBString(m.Genre)
	for _, genre := range genres {
		var genreID int64
		err := s.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM genres
			WHERE name = ?;
			`, genre).Scan(&genreID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := s.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO genres (name)
       			VALUES (?);
       			`, genre)
			if err != nil {
				return err
			}
			genreID, _ = res.LastInsertId()
		}
		_, err = s.DB.Exec( /*sql*/ `INSERT OR IGNORE
       		INTO movies_genres (movie_id, genre_id)
       		VALUES (?, ?);
       		`, m.ImdbID, genreID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) createGenresTx(tx *sql.Tx, m *types.Movie) error {
	genres := util.SplitIMDBString(m.Genre)
	for _, genre := range genres {
		var genreID int64
		err := tx.QueryRow( /*sql*/ `
			SELECT id
			FROM genres
			WHERE name = ?;
			`, genre).Scan(&genreID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := tx.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO genres (name)
       			VALUES (?);
       			`, genre)
			if err != nil {
				return err
			}
			genreID, _ = res.LastInsertId()
		}
		_, err = tx.Exec( /*sql*/ `INSERT OR IGNORE
       		INTO movies_genres (movie_id, genre_id)
       		VALUES (?, ?);
       		`, m.ImdbID, genreID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) createActors(m *types.Movie) error {
	actors := util.SplitIMDBString(m.Actors)
	for _, actor := range actors {
		var actorID int64
		err := s.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM actors
			WHERE name = ?;
			`, actor).Scan(&actorID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := s.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO actors (name)
       			VALUES (?);
       			`, actor)
			if err != nil {
				return err
			}
			actorID, _ = res.LastInsertId()
		}
		_, err = s.DB.Exec( /*sql*/ `
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

func (s *SQLiteStorage) createActorsTx(tx *sql.Tx, m *types.Movie) error {
	actors := util.SplitIMDBString(m.Actors)
	for _, actor := range actors {
		var actorID int64
		err := tx.QueryRow( /*sql*/ `
			SELECT id
			FROM actors
			WHERE name = ?;
			`, actor).Scan(&actorID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := tx.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO actors (name)
       			VALUES (?);
       			`, actor)
			if err != nil {
				return err
			}
			actorID, _ = res.LastInsertId()
		}
		_, err = tx.Exec( /*sql*/ `
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

func (s *SQLiteStorage) CreateUser(username, password string) error {
	hashedPW, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("unable to hash pw: %w", err)
	}
	_, err = s.DB.Exec( /*sql*/ `
		INSERT
		INTO useraccounts (Username, PasswordHash)
		VALUES (?, ?);
	`, username, hashedPW)
	if err != nil {
		return fmt.Errorf("could not create useraccount: %w", err)
	}
	return nil
}

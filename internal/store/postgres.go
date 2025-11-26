package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/jhachmer/gomovie/internal/auth"
	"github.com/jhachmer/gomovie/internal/config"
	"github.com/jhachmer/gomovie/internal/types"
	"github.com/jhachmer/gomovie/internal/util"
	"golang.org/x/crypto/bcrypt"
)

type PostgresStorage struct {
	DB *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		DB: db,
	}
}

func (p *PostgresStorage) Close() error {
	if err := p.DB.Close(); err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) TestDBConnection() error {
	err := p.DB.Ping()
	if err != nil {
		log.Default().Printf("Error pinging database: %v", err)
		return err
	}
	log.Println("connected to DB...")
	return nil
}

func (p *PostgresStorage) CreateAdminAccount(config config.Config) error {
	if config.AdminName == "" || config.AdminPW == "" {
		log.Fatal("admin credentials are not properly set in config!")
		return fmt.Errorf("admin credentials are not properly set in config!")
	}
	hashedPW, err := auth.HashPassword(config.AdminPW)
	if err != nil {
		log.Fatal("error creating admin account")
		return err
	}
	_, err = p.DB.Exec(`--sql
	INSERT OR IGNORE INTO useraccounts (Username, PasswordHash, Active, IsAdmin)
	VALUES (?, ?, ?, ?)
	`, config.AdminName, hashedPW, 1, 1)
	if err != nil {
		return fmt.Errorf("error inserting admin acc %w", err)
	}
	return nil
}

func (p *PostgresStorage) InitDatabaseTables() error {
	// User Accounts
	_, err := p.DB.Exec(`--sql
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
	// Media
	_, err = p.DB.Exec(`--sql
		CREATE TABLE IF NOT EXISTS media (
		id VARCHAR(9) NOT NULL,
		title VARCHAR(255) NOT NULL,
		year VARCHAR(255) NOT NULL,
    	director VARCHAR(500) NOT NULL,
    	runtime VARCHAR(500) NOT NULL,
    	rated VARCHAR(255) NOT NULL,
    	released VARCHAR(500) NOT NULL,
    	plot TEXT NOT NULL,
    	poster VARCHAR(500) NOT NULL,
		seasons VARCHAR(10),
		media_type VARCHAR(255) NOT NULL,
		
		PRIMARY KEY (id));
		`)
	if err != nil {
		return err
	}
	// Ratings
	_, err = p.DB.Exec(`--sql
		CREATE TABLE IF NOT EXISTS ratings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_id VARCHAR(9) NOT NULL,
		source VARCHAR(255) NOT NULL,
		value VARCHAR(50) NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

		FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE);
		`)
	if err != nil {
		return err
	}
	// Entries
	_, err = p.DB.Exec(`--sql
		CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL,
		watched INTEGER DEFAULT 0,
		comment TEXT,
		media_id VARCHAR(9) NOT NULL UNIQUE,
		FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE SET NULL);
		`)
	if err != nil {
		return err
	}
	// Genres
	_, err = p.DB.Exec(`--sql
		CREATE TABLE IF NOT EXISTS genres (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL UNIQUE);
		`)
	if err != nil {
		return err
	}
	// Actors
	_, err = p.DB.Exec(`--sql
		CREATE TABLE IF NOT EXISTS actors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL UNIQUE);
		`)
	if err != nil {
		return err
	}
	// Media Genres MN
	_, err = p.DB.Exec(`--sql
		CREATE TABLE IF NOT EXISTS media_genres (
		media_id VARCHAR(9) NOT NULL,
		genre_id INTEGER NOT NULL,
		PRIMARY KEY (media_id, genre_id),
		FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE,
		FOREIGN KEY (genre_id) REFERENCES genres(id) ON DELETE CASCADE);
		`)
	if err != nil {
		return err
	}
	// Media Actors MN
	_, err = p.DB.Exec(`--sql
		CREATE TABLE IF NOT EXISTS media_actors (
		media_id VARCHAR(9) NOT NULL,
		actor_id INTEGER NOT NULL,
		PRIMARY KEY (media_id, actor_id),
		FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE,
		FOREIGN KEY (actor_id) REFERENCES actors(id) ON DELETE CASCADE);
		`)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) GetWatchCounts() (*types.WatchStats, error) {
	var stats types.WatchStats
	row := p.DB.QueryRow(`--sql
	SELECT
    SUM(CASE WHEN watched = 1 THEN 1 ELSE 0 END) AS watched_count,
    SUM(CASE WHEN watched = 0 THEN 1 ELSE 0 END) AS unwatched_count,
    COUNT(*) AS total_movies
	FROM entries;
	`)
	err := row.Scan(&stats.NumOfWatched, &stats.NumOfUnwatched, &stats.TotalMovies)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (p *PostgresStorage) CheckCredentials(username, password string) (bool, error) {
	var hashedPassword string
	var active bool

	err := p.DB.QueryRow( /*sql*/ `
		SELECT PasswordHash, Active
		FROM UserAccounts
		WHERE Username = ?
		`, username).Scan(&hashedPassword, &active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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

func (p *PostgresStorage) CreateUser(username, password string) error {
	hashedPW, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("unable to hash pw: %w", err)
	}
	_, err = p.DB.Exec( /*sql*/ `
		INSERT
		INTO useraccounts (Username, PasswordHash)
		VALUES (?, ?);
	`, username, hashedPW)
	if err != nil {
		return fmt.Errorf("could not create useraccount: %w", err)
	}
	return nil
}

func (p *PostgresStorage) AdminLoginQuery(username string) (string, error) {
	var passwordHash string
	err := p.DB.QueryRow("SELECT PasswordHash FROM useraccounts WHERE Username = ? AND IsAdmin = 1", username).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

func (p *PostgresStorage) GetUsers() (*sql.Rows, error) {
	rows, err := p.DB.Query("SELECT UserID, Username, Active FROM useraccounts")
	if err != nil {
		return nil, err
	}
	return rows, err
}

func (p *PostgresStorage) ToggleUserActive(active, id int) error {
	_, err := p.DB.Exec("UPDATE useraccounts SET Active = ? WHERE UserID = ?", active, id)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) CreateMovie(m *types.Movie) (*types.Movie, error) {
	_, err := p.DB.Exec( /*sql*/ `
		INSERT INTO media (id, title, year, director, runtime, rated, released, plot, poster, media_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
		`, m.ImdbID, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster, m.Type)
	if err != nil {
		return nil, err
	}
	err = p.createMovieRatings(m)
	if err != nil {
		return nil, err
	}
	err = p.createMovieGenres(m)
	if err != nil {
		return nil, err
	}
	err = p.createMovieActors(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *PostgresStorage) CreateMovieTx(tx *sql.Tx, m *types.Movie) (*types.Movie, error) {
	_, err := tx.Exec( /*sql*/ `
		INSERT INTO media (id, title, year, director, runtime, rated, released, plot, poster, media_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
		`, m.ImdbID, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster, m.Type)
	if err != nil {
		return nil, err
	}
	err = p.createMovieRatingsTx(tx, m)
	if err != nil {
		return nil, err
	}
	err = p.createMovieGenresTx(tx, m)
	if err != nil {
		return nil, err
	}
	err = p.createActorsTx(tx, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *PostgresStorage) UpdateMovie(m *types.Movie) (*types.Movie, error) {
	_, err := p.DB.Exec(`--sql
	UPDATE media
	SET title = ?, year = ?, director = ?, runtime = ?, rated = ?, released = ?, plot = ?, poster = ?
	WHERE id = ?;
	`, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster, m.ImdbID)
	if err != nil {
		return nil, err
	}
	err = p.updateRatings(*m)
	if err != nil {
		return nil, err
	}
	err = p.updateGenres(*m)
	if err != nil {
		return nil, err
	}
	err = p.updateActors(*m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *PostgresStorage) DeleteMedia(imdbId string) error {
	_, err := p.DB.Exec( /*sql*/ `
	DELETE FROM media WHERE id = ?;
	`, imdbId)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) updateRatings(m types.Media) error {
	for _, rating := range m.GetRatings() {
		_, err := p.DB.Exec( /*sql*/ `
			UPDATE ratings
			SET value = ?
			WHERE media_id = ? AND source = ?;
			`, rating.Value, m.GetID(), rating.Source)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) updateGenres(m types.Media) error {
	genres := util.SplitIMDBString(m.GetGenres())

	rows, err := p.DB.Query( /*sql*/ `
	SELECT g.name
	FROM genres g
	INNER JOIN media_genres mg ON g.id = mg.genre_id
	WHERE mg.media_id = ?`, m.GetID())
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
			_, err := p.DB.Exec( /*sql*/ `
			DELETE FROM media_genres
			WHERE
			media_id = ?
			AND
			genre_id = (SELECT id FROM genres WHERE name = ?);
			`, m.GetID(), g)
			if err != nil {
				return err
			}
		}
	}

	for _, genre := range genres {
		var genreID int64
		err := p.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM genres
			WHERE name = ?;
			`, genre).Scan(&genreID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := p.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO genres (name)
       			VALUES (?);
       			`, genre)
			if err != nil {
				return err
			}
			genreID, _ = res.LastInsertId()
		}
		_, err = p.DB.Exec( /*sql*/ `
			INSERT OR IGNORE
       		INTO media_genres (media_id, genre_id)
       		VALUES (?, ?);
       		`, m.GetID(), genreID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) updateActors(m types.Media) error {
	actors := util.SplitIMDBString(m.GetActors())

	rows, err := p.DB.Query( /*sql*/ `
	SELECT a.name
	FROM actors a
	INNER JOIN media_actors ma ON a.id = ma.actor_id
	WHERE ma.media_id = ?`, m.GetID())
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
			_, err := p.DB.Exec( /*sql*/ `
			DELETE FROM media_actors
			WHERE
			media_id = ?
			AND
			actor_id = (SELECT id FROM actors WHERE name = ?);
			`, m.GetID(), a)
			if err != nil {
				return err
			}
		}
	}

	for _, actor := range actors {
		var actorID int64
		err := p.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM actors
			WHERE name = ?;
			`, actor).Scan(&actorID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := p.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO actors (name)
       			VALUES (?);
       			`, actor)
			if err != nil {
				return err
			}
			actorID, _ = res.LastInsertId()
		}
		_, err = p.DB.Exec( /*sql*/ `
			INSERT OR IGNORE
       		INTO media_actors (media_id, actor_id)
       		VALUES (?, ?);
       		`, m.GetID(), actorID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) GetMovieByID(movieID string) (*types.Movie, error) {
	var movie types.Movie

	err := p.DB.QueryRow( /*sql*/ `
        SELECT
            id, title, year, rated, released, runtime, plot, poster, director
        FROM media
        WHERE id = ?`, movieID).Scan(
		&movie.ImdbID, &movie.Title, &movie.Year, &movie.Rated,
		&movie.Released, &movie.Runtime, &movie.Plot, &movie.Poster, &movie.Director)
	if err != nil {
		return nil, err
	}

	rows, err := p.DB.Query( /*sql*/ `
        SELECT g.name
        FROM genres g
        INNER JOIN media_genres mg ON g.id = mg.genre_id
        WHERE mg.media_id = ?`, movieID)
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

	rows, err = p.DB.Query( /*sql*/ `
        SELECT a.name
        FROM actors a
        INNER JOIN media_actors ma ON a.id = ma.actor_id
        WHERE ma.media_id = ?`, movieID)
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

	rows, err = p.DB.Query( /*sql*/ `
        SELECT source, value
        FROM ratings
        WHERE media_id = ?`, movieID)
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

func (p *PostgresStorage) GetAllMovies() ([]*types.MovieInfoData, error) {
	var movies []*types.MovieInfoData
	rows, err := p.DB.Query( /*sql*/ `
        SELECT id
        FROM media
        WHERE media_type = 'movie'
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
		movie, err := p.GetMovieByID(movieID)
		if err != nil {
			return nil, err
		}
		entries, err := p.GetEntries(movieID)
		if err != nil {
			return nil, err
		}
		data := types.MovieInfoData{Movie: movie, Entry: entries}
		movies = append(movies, &data)
	}
	return movies, nil
}

func (p *PostgresStorage) SearchMovie(params types.SearchParams) ([]*types.MovieInfoData, error) {
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
		FROM media m
		LEFT JOIN media_genres mg ON m.id = mg.media_id
		LEFT JOIN genres g ON mg.genre_id = g.id
		LEFT JOIN media_actors ma ON m.id = ma.media_id
		LEFT JOIN actors a ON ma.actor_id = a.id
		`

	if len(filters) > 0 {
		query += "WHERE " + strings.Join(filters, " AND ")
	}

	rows, err := p.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	results := []*types.MovieInfoData{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		mov, err := p.GetMovieByID(id)
		if err != nil {
			return nil, err
		}
		entry, err := p.GetEntries(id)
		if err != nil {
			return nil, err
		}
		results = append(results, &types.MovieInfoData{Movie: mov, Entry: entry})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (p *PostgresStorage) createMovieRatings(m *types.Movie) error {
	for _, rating := range m.Ratings {
		_, err := p.DB.Exec( /*sql*/ `
			INSERT INTO ratings (media_id, source, value)
			VALUES (?, ?, ?);
			`, m.ImdbID, rating.Source, rating.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) createMovieRatingsTx(tx *sql.Tx, m *types.Movie) error {
	for _, rating := range m.Ratings {
		_, err := tx.Exec( /*sql*/ `
			INSERT INTO ratings (media_id, source, value)
			VALUES (?, ?, ?);
			`, m.ImdbID, rating.Source, rating.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) createMovieGenres(m *types.Movie) error {
	genres := util.SplitIMDBString(m.Genre)
	for _, genre := range genres {
		var genreID int64
		err := p.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM genres
			WHERE name = ?;
			`, genre).Scan(&genreID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := p.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO genres (name)
       			VALUES (?);
       			`, genre)
			if err != nil {
				return err
			}
			genreID, _ = res.LastInsertId()
		}
		_, err = p.DB.Exec( /*sql*/ `
			INSERT OR IGNORE
       		INTO media_genres (media_id, genre_id)
       		VALUES (?, ?);
       	`, m.ImdbID, genreID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) createMovieGenresTx(tx *sql.Tx, m *types.Movie) error {
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
       		INTO media_genres (media_id, genre_id)
       		VALUES (?, ?);
       		`, m.ImdbID, genreID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) createMovieActors(m *types.Movie) error {
	actors := util.SplitIMDBString(m.Actors)
	for _, actor := range actors {
		var actorID int64
		err := p.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM actors
			WHERE name = ?;
			`, actor).Scan(&actorID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := p.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO actors (name)
       			VALUES (?);
       			`, actor)
			if err != nil {
				return err
			}
			actorID, _ = res.LastInsertId()
		}
		_, err = p.DB.Exec( /*sql*/ `
			INSERT OR IGNORE
       		INTO media_actors (media_id, actor_id)
       		VALUES (?, ?);
			`, m.ImdbID, actorID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) createActorsTx(tx *sql.Tx, m *types.Movie) error {
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
       		INTO media_actors (media_id, actor_id)
       		VALUES (?, ?);
			`, m.ImdbID, actorID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) CreateSeries(m *types.Series) (*types.Series, error) {
	_, err := p.DB.Exec( /*sql*/ `
		INSERT INTO media (id, title, year, director, runtime, rated, released, plot, poster, media_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
		`, m.ImdbID, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster, m.Type)
	if err != nil {
		return nil, err
	}
	err = p.createSeriesRatings(m)
	if err != nil {
		return nil, err
	}
	err = p.createSeriesGenres(m)
	if err != nil {
		return nil, err
	}
	err = p.createSeriesActors(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *PostgresStorage) createSeriesRatings(m *types.Series) error {
	for _, rating := range m.Ratings {
		_, err := p.DB.Exec( /*sql*/ `
			INSERT INTO ratings (media_id, source, value)
			VALUES (?, ?, ?);
			`, m.ImdbID, rating.Source, rating.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) createSeriesGenres(m *types.Series) error {
	genres := util.SplitIMDBString(m.Genre)
	for _, genre := range genres {
		var genreID int64
		err := p.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM genres
			WHERE name = ?;
			`, genre).Scan(&genreID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := p.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO genres (name)
       			VALUES (?);
       			`, genre)
			if err != nil {
				return err
			}
			genreID, _ = res.LastInsertId()
		}
		_, err = p.DB.Exec( /*sql*/ `
			INSERT OR IGNORE
       		INTO media_genres (media_id, genre_id)
       		VALUES (?, ?);
       	`, m.ImdbID, genreID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) createSeriesActors(m *types.Series) error {
	actors := util.SplitIMDBString(m.Actors)
	for _, actor := range actors {
		var actorID int64
		err := p.DB.QueryRow( /*sql*/ `
			SELECT id
			FROM actors
			WHERE name = ?;
			`, actor).Scan(&actorID)
		if errors.Is(err, sql.ErrNoRows) {
			res, err := p.DB.Exec( /*sql*/ `
				INSERT OR IGNORE
       			INTO actors (name)
       			VALUES (?);
       			`, actor)
			if err != nil {
				return err
			}
			actorID, _ = res.LastInsertId()
		}
		_, err = p.DB.Exec( /*sql*/ `
			INSERT OR IGNORE
       		INTO media_actors (media_id, actor_id)
       		VALUES (?, ?);
			`, m.ImdbID, actorID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) UpdateSeries(m *types.Series) (*types.Series, error) {
	_, err := p.DB.Exec( /*sql*/ `
	UPDATE media
	SET title = ?, year = ?, director = ?, runtime = ?, rated = ?, released = ?, plot = ?, poster = ?
	WHERE id = ?;
	`, m.Title, m.Year, m.Director, m.Runtime, m.Rated, m.Released, m.Plot, m.Poster, m.ImdbID)
	if err != nil {
		return nil, err
	}
	err = p.updateRatings(*m)
	if err != nil {
		return nil, err
	}
	err = p.updateGenres(*m)
	if err != nil {
		return nil, err
	}
	err = p.updateActors(*m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p PostgresStorage) CreateEntry(e *types.Entry, mov *types.Movie) (*types.Entry, error) {
	var exists bool
	row := p.DB.QueryRow( /*sql*/ `
		SELECT EXISTS(SELECT media.title
		FROM media
		WHERE media.id = ?);
		`, mov.ImdbID)
	if err := row.Scan(&exists); err != nil {
		log.Println("movie exists:", exists)
		return nil, err
	} else if !exists {
		_, err := p.CreateMovie(mov)
		if err != nil {
			return nil, err
		}
	}
	var watchedInt = 0
	if e.Watched {
		watchedInt = 1
	}
	res, err := p.DB.Exec( /*sql*/ `
		INSERT INTO entries
		(name, watched, comment, media_id)
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

func (p PostgresStorage) CreateEntryTx(tx *sql.Tx, e *types.Entry, mov *types.Movie) (*types.Entry, error) {
	var exists bool
	row := tx.QueryRow( /*sql*/ `
		SELECT EXISTS(SELECT media.title
		FROM media
		WHERE media.id = ?);
		`, mov.ImdbID)
	if err := row.Scan(&exists); err != nil {
		log.Println("movie exists:", exists)
		return nil, err
	} else if !exists {
		_, err := p.CreateMovieTx(tx, mov)
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
		(name, watched, comment, media_id)
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

func (p PostgresStorage) UpdateEntry(movieId, name, comment string, watched bool) (*types.Entry, error) {
	var watchedInt = 0
	if watched {
		watchedInt = 1
	}
	res, err := p.DB.Exec( /*sql*/ `
		UPDATE entries
		SET name = ?, comment = ?, watched = ?
		WHERE media_id = ?
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

func (p PostgresStorage) DeleteEntry(imdbId string) error {
	_, err := p.DB.Exec( /*sql*/ `
		DELETE FROM entries
		WHERE media_id = ?
	`, imdbId)
	if err != nil {
		return fmt.Errorf("error deleting entry for movie: %s\n%w", imdbId, err)
	}
	return nil
}

func (p PostgresStorage) GetEntries(id string) ([]*types.Entry, error) {
	rows, err := p.DB.Query( /*sql*/ `
		SELECT id, name, watched, comment
		FROM entries
		WHERE media_id = ?;
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

package store

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jhachmer/gomovie/internal/auth"
)

type SQLiteStorage struct {
	DB *sql.DB
}

type Store interface {
	UserStore
	MovieStore
	EntryStore
	StatsStore
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

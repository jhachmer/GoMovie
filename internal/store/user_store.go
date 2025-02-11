package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jhachmer/gomovie/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	CreateUser(username, password string) error
	CheckCredentials(username, password string) (bool, error)
	AdminLoginQuery(username string) (string, error)
	GetUsers() (*sql.Rows, error)
	ToggleUserActive(userID, status int) error
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

func (s *SQLiteStorage) AdminLoginQuery(username string) (string, error) {
	var passwordHash string
	err := s.DB.QueryRow("SELECT PasswordHash FROM useraccounts WHERE Username = ? AND IsAdmin = 1", username).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

func (s *SQLiteStorage) GetUsers() (*sql.Rows, error) {
	rows, err := s.DB.Query("SELECT UserID, Username, Active FROM useraccounts")
	if err != nil {
		return nil, err
	}
	return rows, err
}

func (s *SQLiteStorage) ToggleUserActive(active, id int) error {
	_, err := s.DB.Exec("UPDATE useraccounts SET Active = ? WHERE UserID = ?", active, id)
	if err != nil {
		return err
	}
	return nil
}

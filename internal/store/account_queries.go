package store

import "database/sql"

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
